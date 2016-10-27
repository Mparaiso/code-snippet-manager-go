package smartsnippets

type BeforeEntityDeletedEvent struct {
	Entity
}

type BeforeEntityUpdatedEvent struct {
	Old Entity
	New Entity
}

type BeforeEntityCreatedEvent struct {
	Entity
}

type BeforeResourceCreateEvent struct{}
type AfterResourceCreateEvent struct{}

type BeforeResourceUpdateEvent struct{}
type AfterResourceUpdateEvent struct{}

type BeforeResourceDeleteEvent struct{}
type AfterResourceDeleteEvent struct{}
