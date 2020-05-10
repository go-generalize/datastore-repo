// +build internal

package tests

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	task "github.com/go-generalize/repo_generator/testfiles/a"
)

func initDatastoreClient(t *testing.T) *datastore.Client {
	t.Helper()

	if os.Getenv("DATASTORE_EMULATOR_HOST") == "" {
		os.Setenv("DATASTORE_EMULATOR_HOST", "localhost:8000")
	}

	os.Setenv("DATASTORE_PROJECT_ID", "project-id-in-google")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := datastore.NewClient(ctx, "")

	if err != nil {
		t.Fatalf("failed to initialize datastore client: %+v", err)
	}

	return client
}

func compareTask(t *testing.T, expected, actual *task.Task) {
	t.Helper()

	if actual.ID != expected.ID {
		t.Fatalf("unexpected id: %d(expected: %d)", actual.ID, expected.ID)
	}

	if !actual.Created.Equal(expected.Created) {
		t.Fatalf("unexpected time: %s(expected: %s)", actual.Created, expected.Created)
	}

	if actual.Desc != expected.Desc {
		t.Fatalf("unexpected desc: %s(expected: %s)", actual.Desc, expected.Created)
	}

	if actual.Done != expected.Done {
		t.Fatalf("unexpected done: %v(expected: %v)", actual.Done, expected.Done)
	}
}

func TestDatastoreListTask(t *testing.T) {
	client := initDatastoreClient(t)

	taskRepo := task.NewTaskRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	var ids []int64
	defer func() {
		defer cancel()
		if err := taskRepo.DeleteMultiByIDs(ctx, ids); err != nil {
			t.Fatal(err)
		}
	}()

	now := time.Now()
	desc := "Hello, World!"

	tks := make([]*task.Task, 0)
	for i := int64(1); i <= 10; i++ {
		tk := &task.Task{
			ID:      i * 100,
			Created: now,
			Desc:    fmt.Sprintf("%s%d", desc, i),
			Done:    true,
			Count:   int(i),
		}
		tks = append(tks, tk)
	}
	ids, err := taskRepo.InsertMulti(ctx, tks)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	t.Run("int(1件)", func(t *testing.T) {
		req := &task.TaskListReq{
			Count: task.IntegerCriteria("1"), // FIXME 2 この実装をどうにかしたい
		}

		tasks, err := taskRepo.List(ctx, req, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("bool(10件)", func(t *testing.T) {
		req := &task.TaskListReq{
			Done: task.BoolCriteriaTrue,
		}

		tasks, err := taskRepo.List(ctx, req, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})

	t.Run("time.Time(10件)", func(t *testing.T) {
		req := &task.TaskListReq{
			Created: now,
		}

		tasks, err := taskRepo.List(ctx, req, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})
}

func TestDatastoreListNameWithIndexes(t *testing.T) {
	client := initDatastoreClient(t)

	nameRepo := task.NewNameRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	var ids []int64
	defer func() {
		defer cancel()
		if err := nameRepo.DeleteMultiByIDs(ctx, ids); err != nil {
			t.Fatal(err)
		}
	}()

	now := time.Now()
	desc := "Hello, World!"
	desc2 := "Prefix, Test!"

	tks := make([]*task.Name, 0)
	for i := int64(1); i <= 10; i++ {
		tk := &task.Name{
			ID:      i,
			Created: now,
			Desc:    fmt.Sprintf("%s%d", desc, i),
			Desc2:   fmt.Sprintf("%s%d", desc2, i),
			Done:    true,
			Count:   int(i),
		}
		tks = append(tks, tk)
	}
	ids, err := nameRepo.InsertMulti(ctx, tks)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	t.Run("int(1件)", func(t *testing.T) {
		req := &task.NameListReq{
			Count: task.IntegerCriteria("1"), // FIXME 2 この実装をどうにかしたい
		}

		tasks, err := nameRepo.List(ctx, req, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 1 {
			t.Fatal("not match")
		}
	})

	t.Run("bool(10件)", func(t *testing.T) {
		req := &task.NameListReq{
			Done: task.BoolCriteriaTrue,
		}

		tasks, err := nameRepo.List(ctx, req, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})

	t.Run("like(10件)", func(t *testing.T) {
		req := &task.NameListReq{
			Desc: "ll",
		}

		tasks, err := nameRepo.List(ctx, req, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})

	t.Run("prefix", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			req := &task.NameListReq{
				Desc2: "Pre",
			}

			tasks, err := nameRepo.List(ctx, req, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if len(tasks) != 10 {
				t.Fatal("not match")
			}
		})

		t.Run("Failure", func(t *testing.T) {
			req := &task.NameListReq{
				Desc2: "He",
			}

			tasks, err := nameRepo.List(ctx, req, nil)
			if err != nil {
				t.Fatalf("%+v", err)
			}

			if len(tasks) != 0 {
				t.Fatal("not match")
			}
		})
	})

	t.Run("time.Time(10件)", func(t *testing.T) {
		req := &task.NameListReq{
			Created: now,
		}

		tasks, err := nameRepo.List(ctx, req, nil)
		if err != nil {
			t.Fatalf("%+v", err)
		}

		if len(tasks) != 10 {
			t.Fatal("not match")
		}
	})
}

func TestDatastore(t *testing.T) {
	client := initDatastoreClient(t)

	taskRepo := task.NewTaskRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now()
	desc := "hello"

	id, err := taskRepo.Insert(ctx, &task.Task{
		ID:      1234,
		Desc:    desc,
		Created: now,
		Done:    true,
	})

	if err != nil {
		t.Fatalf("failed to put item: %+v", err)
	}

	ret, err := taskRepo.Get(ctx, id)

	if err != nil {
		t.Fatalf("failed to get item: %+v", err)
	}

	compareTask(t, &task.Task{
		ID:      id,
		Desc:    desc,
		Created: now,
		Done:    true,
	}, ret)

	rets, err := taskRepo.GetMulti(ctx, []int64{id})

	if err != nil {
		t.Fatalf("failed to get item: %+v", err)
	}

	if len(rets) != 1 {
		t.Errorf("GetMulti should return 1 item: %+v", err)
	}

	compareTask(t, &task.Task{
		ID:      id,
		Desc:    desc,
		Created: now,
		Done:    true,
	}, rets[0])

	compareTask(t, &task.Task{
		ID:      id,
		Desc:    desc,
		Created: now,
		Done:    true,
	}, ret)

	if err := taskRepo.DeleteByID(ctx, id); err != nil {
		t.Fatalf("delete failed: %+v", err)
	}

	if _, err := taskRepo.Get(ctx, id); err != datastore.ErrNoSuchEntity {
		t.Fatalf("Get deleted item should return ErrNoSuchEntity: %+v", err)
	}
}
