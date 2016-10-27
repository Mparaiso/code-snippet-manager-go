package smartsnippets

import (
	"fmt"
	"time"

	"github.com/Mparaiso/tiger-go-framework/signal"
)

func BeforeEntityCreatedListener(e signal.Event) error {
	switch event := e.(type) {
	case BeforeEntityCreatedEvent:
		if e, ok := event.Entity.(CreatedUpdatedEntity); ok {
			e.SetCreated(time.Now())
			e.SetUpdated(time.Now())
		}
		if e, ok := event.Entity.(VersionedEntity); ok {
			e.SetVersion(1)
		}
	}
	return nil
}

func BeforeEntityUpdatedListener(e signal.Event) error {
	switch event := e.(type) {
	case BeforeEntityUpdatedEvent:
		if entity, ok := event.Old.(LockedEntity); ok {
			if entity.IsLocked() {
				return fmt.Errorf("Entity is locked and cannot be modified")
			}
		}
		if entity, ok := event.Old.(VersionedEntity); ok {
			if old, new := entity, event.New.(VersionedEntity); old.GetVersion() != new.GetVersion() {
				return fmt.Errorf("Versions do not match old : %d , new : %d", old.GetVersion(), new.GetVersion())
			} else {
				new.SetVersion(old.GetVersion() + 1)
			}
		}
		if entity, ok := event.New.(CreatedUpdatedEntity); ok {
			entity.SetUpdated(time.Now())
		}
		event.New.SetID(event.Old.GetID())
	}
	return nil
}

func BeforeEntityDeletedListener(e signal.Event) error {
	switch event := e.(type) {
	case BeforeEntityDeletedEvent:
		if entity, ok := event.Entity.(LockedEntity); ok {
			if entity.IsLocked() {
				return fmt.Errorf("Entity is locked and cannot be modified")
			}
		}
	}
	return nil
}
