package messages

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Wahab-039/ChatApp/internal/models"
	appmqtt "github.com/Wahab-039/ChatApp/internal/mqtt"
)

type fakeUsers struct {
	byUsername map[string]models.Credentials
	err        error
}

func (f fakeUsers) FindByUsername(_ context.Context, username string) (models.Credentials, error) {
	if f.err != nil {
		return models.Credentials{}, f.err
	}
	user, ok := f.byUsername[username]
	if !ok {
		return models.Credentials{}, models.ErrUserNotFound
	}
	return user, nil
}

type fakeMessages struct {
	byClientID map[string]models.DirectMessage
	byID       map[string]models.DirectMessage
	createErr  error
	created    []models.DirectMessage
	listResult []models.DirectMessage
	listErr    error
}

func (f *fakeMessages) Create(_ context.Context, senderID, recipientID, body, clientMessageID string) (models.DirectMessage, error) {
	if f.createErr != nil {
		return models.DirectMessage{}, f.createErr
	}
	message := models.DirectMessage{
		ID:              "msg-1",
		SenderID:        senderID,
		RecipientID:     recipientID,
		Body:            body,
		ClientMessageID: clientMessageID,
		CreatedAt:       time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC),
	}
	if f.byClientID == nil {
		f.byClientID = map[string]models.DirectMessage{}
	}
	if f.byID == nil {
		f.byID = map[string]models.DirectMessage{}
	}
	f.byClientID[senderID+"|"+clientMessageID] = message
	f.byID[message.ID] = message
	f.created = append(f.created, message)
	return message, nil
}

func (f *fakeMessages) FindBySenderAndClientMessageID(_ context.Context, senderID, clientMessageID string) (models.DirectMessage, error) {
	if f.byClientID == nil {
		return models.DirectMessage{}, models.ErrMessageNotFound
	}
	message, ok := f.byClientID[senderID+"|"+clientMessageID]
	if !ok {
		return models.DirectMessage{}, models.ErrMessageNotFound
	}
	return message, nil
}

func (f *fakeMessages) FindByID(_ context.Context, id string) (models.DirectMessage, error) {
	if f.byID == nil {
		return models.DirectMessage{}, models.ErrMessageNotFound
	}
	message, ok := f.byID[id]
	if !ok {
		return models.DirectMessage{}, models.ErrMessageNotFound
	}
	return message, nil
}

func (f *fakeMessages) ListConversation(_ context.Context, _, _ string, _, _ *models.DirectMessage, limit int) ([]models.DirectMessage, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if f.listResult != nil {
		if len(f.listResult) > limit {
			return f.listResult[:limit], nil
		}
		return f.listResult, nil
	}
	return nil, nil
}

type fakePublisher struct {
	calls []appmqtt.Event
	err   error
}

func (f *fakePublisher) PublishToUserInbox(_ context.Context, _ string, event appmqtt.Event) error {
	f.calls = append(f.calls, event)
	return f.err
}

func TestSendDirectPersistsAndPublishes(t *testing.T) {
	t.Parallel()

	repo := &fakeMessages{}
	publisher := &fakePublisher{}
	service := NewService(fakeUsers{byUsername: map[string]models.Credentials{
		"bob": {User: models.User{ID: "user-b", Username: "bob"}},
	}}, repo, publisher)

	result, err := service.SendDirect(context.Background(), "user-a", "bob", "hello", "client-1")
	if err != nil {
		t.Fatalf("SendDirect() error = %v", err)
	}
	if !result.Created {
		t.Fatal("expected Created=true")
	}
	if len(repo.created) != 1 {
		t.Fatalf("created messages = %d, want 1", len(repo.created))
	}
	if len(publisher.calls) != 1 {
		t.Fatalf("publish calls = %d, want 1", len(publisher.calls))
	}
	if publisher.calls[0].Type != appmqtt.EventTypeMessageNew {
		t.Fatalf("event type = %q", publisher.calls[0].Type)
	}
}

func TestSendDirectIdempotentReplay(t *testing.T) {
	t.Parallel()

	existing := models.DirectMessage{
		ID:              "msg-existing",
		SenderID:        "user-a",
		RecipientID:     "user-b",
		Body:            "hello",
		ClientMessageID: "client-1",
		CreatedAt:       time.Now().UTC(),
	}
	repo := &fakeMessages{byClientID: map[string]models.DirectMessage{
		"user-a|client-1": existing,
	}}
	publisher := &fakePublisher{}
	service := NewService(fakeUsers{byUsername: map[string]models.Credentials{
		"bob": {User: models.User{ID: "user-b", Username: "bob"}},
	}}, repo, publisher)

	result, err := service.SendDirect(context.Background(), "user-a", "bob", "hello", "client-1")
	if err != nil {
		t.Fatalf("SendDirect() error = %v", err)
	}
	if result.Created {
		t.Fatal("expected Created=false for idempotent replay")
	}
	if result.Message.ID != existing.ID {
		t.Fatalf("message id = %q, want %q", result.Message.ID, existing.ID)
	}
	if len(publisher.calls) != 0 {
		t.Fatalf("expected no republish on idempotent replay, got %d", len(publisher.calls))
	}
}

func TestSendDirectRejectsSelf(t *testing.T) {
	t.Parallel()

	service := NewService(fakeUsers{byUsername: map[string]models.Credentials{
		"wahab": {User: models.User{ID: "user-a", Username: "wahab"}},
	}}, &fakeMessages{}, &fakePublisher{})

	_, err := service.SendDirect(context.Background(), "user-a", "wahab", "hello", "client-1")
	if !errors.Is(err, ErrCannotMessageSelf) {
		t.Fatalf("error = %v, want %v", err, ErrCannotMessageSelf)
	}
}

func TestSendDirectPublishFailureStillSucceeds(t *testing.T) {
	t.Parallel()

	service := NewService(fakeUsers{byUsername: map[string]models.Credentials{
		"bob": {User: models.User{ID: "user-b", Username: "bob"}},
	}}, &fakeMessages{}, &fakePublisher{err: errors.New("broker down")})

	result, err := service.SendDirect(context.Background(), "user-a", "bob", "hello", "client-1")
	if err != nil {
		t.Fatalf("SendDirect() error = %v", err)
	}
	if !result.Created {
		t.Fatal("expected Created=true even when publish fails")
	}
}

func TestListDirectReturnsConversationPage(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 21, 12, 0, 0, 0, time.UTC)
	repo := &fakeMessages{
		listResult: []models.DirectMessage{
			{ID: "m1", SenderID: "user-a", RecipientID: "user-b", Body: "hi", CreatedAt: now},
			{ID: "m2", SenderID: "user-b", RecipientID: "user-a", Body: "hey", CreatedAt: now.Add(time.Minute)},
		},
	}
	service := NewService(fakeUsers{byUsername: map[string]models.Credentials{
		"bob": {User: models.User{ID: "user-b", Username: "bob"}},
	}}, repo, &fakePublisher{})

	result, err := service.ListDirect(context.Background(), "user-a", HistoryQuery{PeerUsername: "bob", Limit: 50})
	if err != nil {
		t.Fatalf("ListDirect() error = %v", err)
	}
	if len(result.Messages) != 2 {
		t.Fatalf("messages = %d, want 2", len(result.Messages))
	}
	if result.NextAfter != "m2" {
		t.Fatalf("next_after = %q, want m2", result.NextAfter)
	}
}

func TestListDirectRejectsMissingPeer(t *testing.T) {
	t.Parallel()

	service := NewService(fakeUsers{}, &fakeMessages{}, &fakePublisher{})
	_, err := service.ListDirect(context.Background(), "user-a", HistoryQuery{})
	if !errors.Is(err, ErrPeerRequired) {
		t.Fatalf("error = %v, want %v", err, ErrPeerRequired)
	}
}
