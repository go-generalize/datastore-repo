// +build internal

package tests

import (
	"context"
	"os"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	task "github.com/go-generalize/repo_generator/testfiles/c"
	"github.com/google/uuid"
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

	if !actual.ID.Equal(expected.ID) {
		t.Fatalf("unexpected id: %s(expected: %s)", actual.ID, expected.ID)
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

func TestDatastore(t *testing.T) {
	client := initDatastoreClient(t)

	taskRepo := task.NewTaskRepository(client)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Unix(time.Now().Unix(), 0)
	desc := "hello"

	id, err := taskRepo.Insert(ctx, &task.Task{
		ID:      datastore.NameKey("Task", uuid.New().String(), nil),
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

	rets, err := taskRepo.GetMulti(ctx, []*datastore.Key{id})

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
