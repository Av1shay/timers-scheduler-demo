// Code generated by ent, DO NOT EDIT.

package ent

import (
	"github.com/Av1shay/timers-scheduler-demo/ent/schema"
	"github.com/Av1shay/timers-scheduler-demo/ent/task"
	"github.com/Av1shay/timers-scheduler-demo/ent/taskhistory"
	"time"
)

// The init function reads all schema descriptors with runtime code
// (default values, validators, hooks and policies) and stitches it
// to their package variables.
func init() {
	taskFields := schema.Task{}.Fields()
	_ = taskFields
	// taskDescCreatedAt is the schema descriptor for created_at field.
	taskDescCreatedAt := taskFields[3].Descriptor()
	// task.DefaultCreatedAt holds the default value on creation for the created_at field.
	task.DefaultCreatedAt = taskDescCreatedAt.Default.(func() time.Time)
	// taskDescUpdatedAt is the schema descriptor for updated_at field.
	taskDescUpdatedAt := taskFields[4].Descriptor()
	// task.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	task.DefaultUpdatedAt = taskDescUpdatedAt.Default.(func() time.Time)
	// task.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	task.UpdateDefaultUpdatedAt = taskDescUpdatedAt.UpdateDefault.(func() time.Time)
	taskhistoryFields := schema.TaskHistory{}.Fields()
	_ = taskhistoryFields
	// taskhistoryDescCreatedAt is the schema descriptor for created_at field.
	taskhistoryDescCreatedAt := taskhistoryFields[1].Descriptor()
	// taskhistory.DefaultCreatedAt holds the default value on creation for the created_at field.
	taskhistory.DefaultCreatedAt = taskhistoryDescCreatedAt.Default.(func() time.Time)
	// taskhistoryDescUpdatedAt is the schema descriptor for updated_at field.
	taskhistoryDescUpdatedAt := taskhistoryFields[2].Descriptor()
	// taskhistory.DefaultUpdatedAt holds the default value on creation for the updated_at field.
	taskhistory.DefaultUpdatedAt = taskhistoryDescUpdatedAt.Default.(func() time.Time)
	// taskhistory.UpdateDefaultUpdatedAt holds the default value on update for the updated_at field.
	taskhistory.UpdateDefaultUpdatedAt = taskhistoryDescUpdatedAt.UpdateDefault.(func() time.Time)
}
