// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"github.com/Av1shay/timers-scheduler-demo/ent/predicate"
	"github.com/Av1shay/timers-scheduler-demo/ent/taskhistory"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
)

// TaskHistoryDelete is the builder for deleting a TaskHistory entity.
type TaskHistoryDelete struct {
	config
	hooks    []Hook
	mutation *TaskHistoryMutation
}

// Where appends a list predicates to the TaskHistoryDelete builder.
func (thd *TaskHistoryDelete) Where(ps ...predicate.TaskHistory) *TaskHistoryDelete {
	thd.mutation.Where(ps...)
	return thd
}

// Exec executes the deletion query and returns how many vertices were deleted.
func (thd *TaskHistoryDelete) Exec(ctx context.Context) (int, error) {
	var (
		err      error
		affected int
	)
	if len(thd.hooks) == 0 {
		affected, err = thd.sqlExec(ctx)
	} else {
		var mut Mutator = MutateFunc(func(ctx context.Context, m Mutation) (Value, error) {
			mutation, ok := m.(*TaskHistoryMutation)
			if !ok {
				return nil, fmt.Errorf("unexpected mutation type %T", m)
			}
			thd.mutation = mutation
			affected, err = thd.sqlExec(ctx)
			mutation.done = true
			return affected, err
		})
		for i := len(thd.hooks) - 1; i >= 0; i-- {
			if thd.hooks[i] == nil {
				return 0, fmt.Errorf("ent: uninitialized hook (forgotten import ent/runtime?)")
			}
			mut = thd.hooks[i](mut)
		}
		if _, err := mut.Mutate(ctx, thd.mutation); err != nil {
			return 0, err
		}
	}
	return affected, err
}

// ExecX is like Exec, but panics if an error occurs.
func (thd *TaskHistoryDelete) ExecX(ctx context.Context) int {
	n, err := thd.Exec(ctx)
	if err != nil {
		panic(err)
	}
	return n
}

func (thd *TaskHistoryDelete) sqlExec(ctx context.Context) (int, error) {
	_spec := &sqlgraph.DeleteSpec{
		Node: &sqlgraph.NodeSpec{
			Table: taskhistory.Table,
			ID: &sqlgraph.FieldSpec{
				Type:   field.TypeInt,
				Column: taskhistory.FieldID,
			},
		},
	}
	if ps := thd.mutation.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	affected, err := sqlgraph.DeleteNodes(ctx, thd.driver, _spec)
	if err != nil && sqlgraph.IsConstraintError(err) {
		err = &ConstraintError{msg: err.Error(), wrap: err}
	}
	return affected, err
}

// TaskHistoryDeleteOne is the builder for deleting a single TaskHistory entity.
type TaskHistoryDeleteOne struct {
	thd *TaskHistoryDelete
}

// Exec executes the deletion query.
func (thdo *TaskHistoryDeleteOne) Exec(ctx context.Context) error {
	n, err := thdo.thd.Exec(ctx)
	switch {
	case err != nil:
		return err
	case n == 0:
		return &NotFoundError{taskhistory.Label}
	default:
		return nil
	}
}

// ExecX is like Exec, but panics if an error occurs.
func (thdo *TaskHistoryDeleteOne) ExecX(ctx context.Context) {
	thdo.thd.ExecX(ctx)
}