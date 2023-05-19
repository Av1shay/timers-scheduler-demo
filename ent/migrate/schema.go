// Code generated by ent, DO NOT EDIT.

package migrate

import (
	"entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/schema/field"
)

var (
	// TasksColumns holds the columns for the "tasks" table.
	TasksColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "due_date", Type: field.TypeTime},
		{Name: "webhook_url", Type: field.TypeString},
		{Name: "status", Type: field.TypeEnum, Enums: []string{"pending", "running", "done"}, Default: "pending"},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
	}
	// TasksTable holds the schema information for the "tasks" table.
	TasksTable = &schema.Table{
		Name:       "tasks",
		Columns:    TasksColumns,
		PrimaryKey: []*schema.Column{TasksColumns[0]},
		Indexes: []*schema.Index{
			{
				Name:    "task_due_date",
				Unique:  false,
				Columns: []*schema.Column{TasksColumns[1]},
			},
			{
				Name:    "task_status",
				Unique:  false,
				Columns: []*schema.Column{TasksColumns[3]},
			},
			{
				Name:    "task_due_date_status",
				Unique:  false,
				Columns: []*schema.Column{TasksColumns[1], TasksColumns[3]},
			},
		},
	}
	// TaskHistoriesColumns holds the columns for the "task_histories" table.
	TaskHistoriesColumns = []*schema.Column{
		{Name: "id", Type: field.TypeInt, Increment: true},
		{Name: "error", Type: field.TypeString, Nullable: true},
		{Name: "created_at", Type: field.TypeTime},
		{Name: "updated_at", Type: field.TypeTime},
		{Name: "task_histories", Type: field.TypeInt, Nullable: true},
	}
	// TaskHistoriesTable holds the schema information for the "task_histories" table.
	TaskHistoriesTable = &schema.Table{
		Name:       "task_histories",
		Columns:    TaskHistoriesColumns,
		PrimaryKey: []*schema.Column{TaskHistoriesColumns[0]},
		ForeignKeys: []*schema.ForeignKey{
			{
				Symbol:     "task_histories_tasks_histories",
				Columns:    []*schema.Column{TaskHistoriesColumns[4]},
				RefColumns: []*schema.Column{TasksColumns[0]},
				OnDelete:   schema.SetNull,
			},
		},
	}
	// Tables holds all the tables in the schema.
	Tables = []*schema.Table{
		TasksTable,
		TaskHistoriesTable,
	}
)

func init() {
	TaskHistoriesTable.ForeignKeys[0].RefTable = TasksTable
}
