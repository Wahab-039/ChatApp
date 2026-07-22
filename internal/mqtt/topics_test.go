package mqtt

import "testing"

func TestUserInboxTopic(t *testing.T) {
	t.Parallel()

	got := UserInboxTopic("1d3e30a1-0c19-411c-b0cf-034c6ae88603")
	want := "chat/user/1d3e30a1-0c19-411c-b0cf-034c6ae88603/inbox"
	if got != want {
		t.Fatalf("UserInboxTopic() = %q, want %q", got, want)
	}
}
