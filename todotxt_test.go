package todotxt

import (
        "testing"
        "github.com/stretchr/testify/assert"
)

func TestLoadTaskList (t *testing.T) {
        tasklist := LoadTaskList("todo.txt")
        assert.Equal(t, tasklist.Len(), 8, "Something went wrong with LoadTaskList")
}

func TestLoadTaskListNonExistent (t *testing.T) {
        defer func(){
                if r:=recover(); r!=nil {
                        // recovered
                } else {
                        t.Errorf("Something went seriously wrong")
                }
        }()
        tasklist := LoadTaskList("nonexistent-file.txt")

        t.Errorf("Something is still wrong %v", tasklist)
}

func TestCreateTask (t *testing.T) {
        task := CreateTask(1, "(A) +funny task with prioity and project")

        assert.Equal(t, task.id, 1, "id should be 1")
        assert.Equal(t, rune(task.priority), rune('A'), "priority should be A")

        projects := make([]string, 1)
        projects[0] = "+funny"

        assert.Equal(t, task.projects, projects, "there should be a project for sure")
        assert.Equal(t, task.todo, "+funny task with prioity and project", "todo should equal")

}
