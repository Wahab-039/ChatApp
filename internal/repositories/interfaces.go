package repositories

// This file contains repository-level interface contracts.
// Currently, repositories implement interfaces defined in each services subpackage
// (e.g. users.UserRepository, messages.MessageRepository) rather than defining
// their own, following the dependency inversion principle.
//
// If repositories need to depend on other persistence abstractions in the future,
// those interfaces would be defined here.
