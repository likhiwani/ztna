package command

import (
	"ztna-core/ztna/common/pb/cmd_pb"
	"ztna-core/ztna/controller/change"
	"ztna-core/ztna/controller/fields"
	"ztna-core/ztna/controller/models"
	"ztna-core/ztna/logtrace"

	"github.com/openziti/storage/boltz"
	"github.com/pkg/errors"
)

// EntityMarshaller instances can marshal and unmarshal entities of the type that they manage
// as well as knowing their entity type
type EntityMarshaller[T any] interface {
	// GetEntityTypeId returns the entity type id. This is distinct from the Store entity id
	// which may be shared by types, such as service and router. The entity type is unique
	// for each type
	GetEntityTypeId() string

	// Marshall marshals the entity to a bytes encoded format
	Marshall(entity T) ([]byte, error)

	// Unmarshall unmarshalls the bytes back into an entity
	Unmarshall(bytes []byte) (T, error)
}

// EntityCreator instances can apply a create entity command to create entities of a given type
type EntityCreator[T models.Entity] interface {
	EntityMarshaller[T]

	// ApplyCreate creates the entity described by the given command
	ApplyCreate(cmd *CreateEntityCommand[T], ctx boltz.MutateContext) error
}

// EntityUpdater instances can apply an update entity command to update entities of a given type
type EntityUpdater[T models.Entity] interface {
	EntityMarshaller[T]

	// ApplyUpdate updates the entity described by the given command
	ApplyUpdate(cmd *UpdateEntityCommand[T], ctx boltz.MutateContext) error
}

// EntityDeleter instances can apply a delete entity command to delete entities of a given type
type EntityDeleter interface {
	GetEntityTypeId() string

	// ApplyDelete deletes the entity described by the given command
	ApplyDelete(cmd *DeleteEntityCommand, ctx boltz.MutateContext) error
}

// EntityManager instances can handle create, update and delete entities of a specific type
type EntityManager[T models.Entity] interface {
	EntityCreator[T]
	EntityUpdater[T]
	EntityDeleter
}

type CreateEntityCommand[T models.Entity] struct {
	Context        *change.Context
	Creator        EntityCreator[T]
	Entity         T
	PostCreateHook func(ctx boltz.MutateContext, entity T) error
	Flags          uint32
}

func (self *CreateEntityCommand[T]) Apply(ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.Creator.ApplyCreate(self, ctx)
}

func (self *CreateEntityCommand[T]) Encode() ([]byte, error) {
	logtrace.LogWithFunctionName()
	entityType := self.Creator.GetEntityTypeId()
	encodedEntity, err := self.Creator.Marshall(self.Entity)
	if err != nil {
		return nil, errors.Wrapf(err, "error mashalling entity of type %T (%v)", self.Entity, entityType)
	}
	return cmd_pb.EncodeProtobuf(&cmd_pb.CreateEntityCommand{
		Ctx:        self.Context.ToProtoBuf(),
		EntityType: entityType,
		EntityData: encodedEntity,
		Flags:      self.Flags,
	})
}

func (self *CreateEntityCommand[T]) GetChangeContext() *change.Context {
	logtrace.LogWithFunctionName()
	return self.Context
}

type UpdateEntityCommand[T models.Entity] struct {
	Context       *change.Context
	Updater       EntityUpdater[T]
	Entity        T
	UpdatedFields fields.UpdatedFields
	Flags         uint32
}

func (self *UpdateEntityCommand[T]) Apply(ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.Updater.ApplyUpdate(self, ctx)
}

func (self *UpdateEntityCommand[T]) Encode() ([]byte, error) {
	logtrace.LogWithFunctionName()
	entityType := self.Updater.GetEntityTypeId()
	encodedEntity, err := self.Updater.Marshall(self.Entity)
	if err != nil {
		return nil, errors.Wrapf(err, "error mashalling entity of type %T (%v)", self.Entity, entityType)
	}

	updatedFields, err := fields.UpdatedFieldsToSlice(self.UpdatedFields)
	if err != nil {
		return nil, err
	}

	return cmd_pb.EncodeProtobuf(&cmd_pb.UpdateEntityCommand{
		Ctx:           self.Context.ToProtoBuf(),
		EntityType:    entityType,
		EntityData:    encodedEntity,
		UpdatedFields: updatedFields,
		Flags:         self.Flags,
	})
}

type DeleteEntityCommand struct {
	Context *change.Context
	Deleter EntityDeleter
	Id      string
}

func (self *UpdateEntityCommand[T]) GetChangeContext() *change.Context {
	logtrace.LogWithFunctionName()
	return self.Context
}

func (self *DeleteEntityCommand) Apply(ctx boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.Deleter.ApplyDelete(self, ctx)
}

func (self *DeleteEntityCommand) Encode() ([]byte, error) {
	logtrace.LogWithFunctionName()
	return cmd_pb.EncodeProtobuf(&cmd_pb.DeleteEntityCommand{
		Ctx:        self.Context.ToProtoBuf(),
		EntityId:   self.Id,
		EntityType: self.Deleter.GetEntityTypeId(),
	})
}

func (self *DeleteEntityCommand) GetChangeContext() *change.Context {
	logtrace.LogWithFunctionName()
	return self.Context
}

type SyncSnapshotCommand struct {
	SnapshotId   string
	Snapshot     []byte
	SnapshotSink func(cmd *SyncSnapshotCommand) error
}

func (self *SyncSnapshotCommand) Apply(boltz.MutateContext) error {
	logtrace.LogWithFunctionName()
	return self.SnapshotSink(self)
}

func (self *SyncSnapshotCommand) Encode() ([]byte, error) {
	logtrace.LogWithFunctionName()
	return cmd_pb.EncodeProtobuf(&cmd_pb.SyncSnapshotCommand{
		SnapshotId: self.SnapshotId,
		Snapshot:   self.Snapshot,
	})
}

func (self *SyncSnapshotCommand) GetChangeContext() *change.Context {
	logtrace.LogWithFunctionName()
	return nil
}
