package tests

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/ravendb/ravendb-go-client/examples/northwind"

	ravendb "github.com/ravendb/ravendb-go-client"

	"github.com/stretchr/testify/assert"
)

func assertIllegalArgumentError(t *testing.T, err error, s ...string) {
	assert.Error(t, err)
	if err != nil {
		_, ok := err.(*ravendb.IllegalArgumentError)
		if !ok {
			assert.True(t, ok, "expected error of type *ravendb.IllegalArgumentError, got %T", err)
			return
		}
		if len(s) > 0 {
			panicIf(len(s) > 1, "only 0 or 1 strings are expected as s")
			assert.Equal(t, s[0], err.Error())
		}
	}
}

func assertIllegalStateError(t *testing.T, err error, s ...string) {
	assert.Error(t, err)
	if err != nil {
		_, ok := err.(*ravendb.IllegalStateError)
		if !ok {
			assert.True(t, ok, "expected error of type *ravendb.IllegalStateError, got %T", err)
			return
		}
		if len(s) > 0 {
			panicIf(len(s) > 1, "only 0 or 1 strings are expected as s")
			assert.Equal(t, s[0], err.Error())
		}
	}
}

func goTest(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	session := openSessionMust(t, store)
	user := User{}

	// check validation of arguments to Store and Delete

	{
		// can't store/delete etc. nil
		var v interface{}
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be nil")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.Delete(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. nil pointer
		var v *User
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity of type *tests.User can't be nil")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.Delete(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. struct
		v := user
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type tests.User, try passing *tests.User")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.Delete(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. **struct (double pointer values)
		ptrUser := &user
		v := &ptrUser
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type **tests.User, try passing *tests.User")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.Delete(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		// can't store/delete etc. a map
		var v map[string]interface{}
		err = session.Store(v)
		assertIllegalArgumentError(t, err, "entity can't be of type map[string]interface {}, try passing *map[string]interface {}")
		err = session.StoreWithID(v, "users/1")
		assertIllegalArgumentError(t, err)
		err = session.Delete(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetMetadataFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetChangeVectorFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.GetLastModifiedFor(v)
		assertIllegalArgumentError(t, err)
		_, err = session.HasChanged(v)
		assertIllegalArgumentError(t, err)
		err = session.Evict(v)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Patch(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().Increment(v, "foo", 1)
		assertIllegalArgumentError(t, err)
		err = session.Advanced().PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err)
		err = session.Refresh(v)
		assertIllegalArgumentError(t, err)
	}

	{
		v := &User{} // dummy value that only has to pass type check
		adv := session.Advanced()

		err = adv.Increment(v, "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.Increment(v, "foo", nil)
		assertIllegalArgumentError(t, err, "valueToAdd can't be nil")

		err = adv.IncrementByID("", "foo", 1)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.IncrementByID("id", "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.IncrementByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "valueToAdd can't be nil")

		err = adv.Patch(v, "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.Patch(v, "foo", nil)
		assertIllegalArgumentError(t, err, "value can't be nil")

		err = adv.PatchByID("", "foo", 1)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.PatchByID("id", "", 1)
		assertIllegalArgumentError(t, err, "path can't be empty string")
		err = adv.PatchByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "value can't be nil")

		err = adv.PatchArray(v, "", nil)
		assertIllegalArgumentError(t, err, "pathToArray can't be empty string")
		err = adv.PatchArray(v, "foo", nil)
		assertIllegalArgumentError(t, err, "arrayAdder can't be nil")

		err = adv.PatchArrayByID("", "foo", nil)
		assertIllegalArgumentError(t, err, "id can't be empty string")
		err = adv.PatchArrayByID("id", "", nil)
		assertIllegalArgumentError(t, err, "pathToArray can't be empty string")
		err = adv.PatchArrayByID("id", "foo", nil)
		assertIllegalArgumentError(t, err, "arrayAdder can't be nil")
	}

	{
		_, err = session.Exists("")
		assertIllegalArgumentError(t, err, "id cannot be empty string")
	}

	session.Close()
}

func goStore(t *testing.T, session *ravendb.DocumentSession) []*User {
	logTestName()

	var err error
	var res []*User
	{
		names := []string{"John", "Mary", "Paul"}
		for _, name := range names {
			u := &User{}
			u.setName(name)
			err := session.Store(u)
			assert.NoError(t, err)
			res = append(res, u)
		}
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
	}
	return res
}

func goTestGetLastModifiedForAndChanges(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error
	var changed, hasChanges bool

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	var users []*User
	var lastModifiedFirst *time.Time
	{
		session := openSessionMust(t, store)
		users = goStore(t, session)
		lastModifiedFirst, err = session.GetLastModifiedFor(users[0])
		assert.NoError(t, err)
		assert.NotNil(t, lastModifiedFirst)
		session.Close()
	}

	{
		session := openSessionMust(t, store)

		// test HasChanges()
		hasChanges = session.HasChanges()
		assert.False(t, hasChanges)

		var u *User
		id := users[0].ID
		err = session.Load(&u, id)
		assert.NoError(t, err)
		assert.Equal(t, id, u.ID)
		lastModified, err := session.GetLastModifiedFor(u)
		assert.NoError(t, err)
		assert.Equal(t, *lastModifiedFirst, *lastModified)

		changed, err = session.HasChanged(u)
		assert.NoError(t, err)
		assert.False(t, changed)

		// check last modified changes after modification
		u.Age = 5
		err = session.Store(u)
		assert.NoError(t, err)

		changed, err = session.HasChanged(u)
		assert.NoError(t, err)
		assert.True(t, changed)

		hasChanges = session.HasChanges()
		assert.True(t, hasChanges)

		err = session.SaveChanges()
		assert.NoError(t, err)

		lastModified, err = session.GetLastModifiedFor(u)
		assert.NoError(t, err)
		diff := (*lastModified).Sub(*lastModifiedFirst)
		assert.True(t, diff > 0)

		session.Close()
	}

	{
		// test HasChanged() detects deletion
		session := openSessionMust(t, store)
		var u *User
		id := users[0].ID
		err = session.Load(&u, id)
		assert.NoError(t, err)

		err = session.Delete(u)
		assert.NoError(t, err)

		/*
			// TODO: should deleted items be reported as changed?
			changed, err = session.HasChanged(u)
			assert.NoError(t, err)
			assert.True(t, changed)
		*/

		hasChanges = session.HasChanges()
		assert.True(t, hasChanges)

		// Evict undoes deletion so we shouldn't have changes
		err = session.Evict(u)
		assert.NoError(t, err)

		changed, err = session.HasChanged(u)
		assert.NoError(t, err)
		assert.False(t, changed)

		hasChanges = session.HasChanges()
		assert.False(t, hasChanges)
	}
}

func goTestListeners(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	nBeforeStoreCalledCount := 0
	beforeStore := func(event *ravendb.BeforeStoreEventArgs) {
		_, ok := event.Entity.(*User)
		assert.True(t, ok)
		nBeforeStoreCalledCount++
	}
	beforeStoreID := store.AddBeforeStoreListener(beforeStore)

	nAfterSaveChangesCalledCount := 0
	afterSaveChanges := func(event *ravendb.AfterSaveChangesEventArgs) {
		_, ok := event.Entity.(*User)
		assert.True(t, ok)
		nAfterSaveChangesCalledCount++
	}
	afterSaveChangesID := store.AddAfterSaveChangesListener(afterSaveChanges)

	nBeforeDeleteCalledCount := 0
	beforeDelete := func(event *ravendb.BeforeDeleteEventArgs) {
		u, ok := event.Entity.(*User)
		assert.True(t, ok)
		assert.Equal(t, "users/1-A", u.ID)
		nBeforeDeleteCalledCount++
	}
	beforeDeleteID := store.AddBeforeDeleteListener(beforeDelete)

	nBeforeQueryCalledCount := 0
	beforeQuery := func(event *ravendb.BeforeQueryEventArgs) {
		nBeforeQueryCalledCount++
	}
	beforeQueryID := store.AddBeforeQueryListener(beforeQuery)

	{
		assert.Equal(t, 0, nBeforeStoreCalledCount)
		assert.Equal(t, 0, nAfterSaveChangesCalledCount)
		session := openSessionMust(t, store)
		users := goStore(t, session)
		session.Close()
		assert.Equal(t, len(users), nBeforeStoreCalledCount)
		assert.Equal(t, len(users), nAfterSaveChangesCalledCount)
	}

	{
		assert.Equal(t, 0, nBeforeDeleteCalledCount)
		session := openSessionMust(t, store)
		var u *User
		err = session.Load(&u, "users/1-A")
		assert.NoError(t, err)
		err = session.Delete(u)
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()
		assert.Equal(t, 1, nBeforeDeleteCalledCount)
	}

	{
		assert.Equal(t, 0, nBeforeQueryCalledCount)
		session := openSessionMust(t, store)
		tp := reflect.TypeOf(&User{})
		q := session.QueryCollectionForType(tp)
		var users []*User
		err = q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		session.Close()
		assert.Equal(t, 1, nBeforeQueryCalledCount)
	}

	store.RemoveBeforeStoreListener(beforeStoreID)
	store.RemoveAfterSaveChangesListener(afterSaveChangesID)
	store.RemoveBeforeDeleteListener(beforeDeleteID)
	store.RemoveBeforeQueryListener(beforeQueryID)

	{
		// verify those listeners were removed
		nBeforeStoreCalledCountPrev := nBeforeStoreCalledCount
		nAfterSaveChangesCalledCountPrev := nAfterSaveChangesCalledCount
		nBeforeDeleteCalledCountPrev := nBeforeDeleteCalledCount
		nBeforeQueryCalledCountPrev := nBeforeQueryCalledCount

		session := openSessionMust(t, store)

		var users []*User
		q := session.QueryCollectionForType(userType)
		err = q.GetResults(&users)
		assert.NoError(t, err)
		assert.Equal(t, 2, len(users))
		assert.Equal(t, nBeforeQueryCalledCountPrev, nBeforeQueryCalledCount)

		u := &User{}
		err = session.Store(u)
		assert.NoError(t, err)
		err = session.DeleteByID("users/2-A", "")
		assert.NoError(t, err)
		err = session.SaveChanges()
		assert.NoError(t, err)
		session.Close()

		assert.Equal(t, nBeforeStoreCalledCountPrev, nBeforeStoreCalledCount)
		assert.Equal(t, nAfterSaveChangesCalledCountPrev, nAfterSaveChangesCalledCount)
		assert.Equal(t, nBeforeDeleteCalledCountPrev, nBeforeDeleteCalledCount)
	}

	{
		// test that Refresh() only works if entity is in session
		session := openSessionMust(t, store)
		var u *User
		err = session.Load(&u, "users/3-A")
		assert.NoError(t, err)
		assert.NotNil(t, u)
		err = session.Refresh(u)
		assert.NoError(t, err)

		err = session.Refresh(u)
		assert.NoError(t, err)

		for i := 0; err == nil && i < 32; i++ {
			err = session.Refresh(u)
		}
		assertIllegalStateError(t, err, "exceeded max number of requests per session of 32")

		session.Close()
	}

	{
		// check Load() does proper argument validation
		session := openSessionMust(t, store)

		var v *User
		err = session.Load(&v, "")
		assertIllegalArgumentError(t, err, "id cannot be empty string")

		err = session.Load(nil, "id")
		assertIllegalArgumentError(t, err, "result can't be nil")

		err = session.Load(User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type tests.User, try passing **tests.User")

		err = session.Load(&User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type *tests.User, try passing **tests.User")

		err = session.Load([]*User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type []*tests.User")

		err = session.Load(&[]*User{}, "id")
		assertIllegalArgumentError(t, err, "result can't be of type *[]*tests.User")

		var n int
		err = session.Load(n, "id")
		assertIllegalArgumentError(t, err, "result can't be of type int")
		err = session.Load(&n, "id")
		assertIllegalArgumentError(t, err, "result can't be of type *int")
		nPtr := &n
		err = session.Load(&nPtr, "id")
		assertIllegalArgumentError(t, err, "result can't be of type **int")

		session.Close()
	}

	{
		// check LoadMulti() does proper argument validation
		session := openSessionMust(t, store)

		var v map[string]*User
		err = session.LoadMulti(v, nil)
		assertIllegalArgumentError(t, err, "ids cannot be empty array")
		err = session.LoadMulti(&v, []string{})
		assertIllegalArgumentError(t, err, "ids cannot be empty array")

		err = session.LoadMulti(User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type tests.User, must be map[string]<type>")

		err = session.LoadMulti(&User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type *tests.User, must be map[string]<type>")

		err = session.LoadMulti(map[int]*User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[int]*tests.User, must be map[string]<type>")

		err = session.LoadMulti(map[string]int{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[string]int, must be map[string]<type>")

		err = session.LoadMulti(map[string]*int{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[string]*int, must be map[string]<type>")

		err = session.LoadMulti(map[string]User{}, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be of type map[string]tests.User, must be map[string]<type>")

		err = session.LoadMulti(v, []string{"id"})
		assertIllegalArgumentError(t, err, "results can't be a nil map")

		session.Close()
	}

}

// TODO: this must be more comprehensive. Need to test all APIs.
func goTestStoreMap(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error

	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		m := map[string]interface{}{
			"foo":     5,
			"bar":     true,
			"nullVal": nil,
			"strVal":  "a string",
		}
		err = session.StoreWithID(&m, "maps/1")
		assert.NoError(t, err)

		m2 := map[string]interface{}{
			"foo":    8,
			"strVal": "more string",
		}
		err = session.Store(&m2)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		meta, err := session.GetMetadataFor(m)
		assertIllegalArgumentError(t, err, "instance can't be of type map[string]interface {}, try passing *map[string]interface {}")
		assert.Nil(t, meta)

		meta, err = session.GetMetadataFor(&m)
		assert.NoError(t, err)
		assert.NotNil(t, meta)

		session.Close()
	}

	{
		session := openSessionMust(t, store)
		var mp *map[string]interface{}
		err = session.Load(&mp, "maps/1")
		assert.NoError(t, err)
		m := *mp
		assert.Equal(t, float64(5), m["foo"])
		assert.Equal(t, "a string", m["strVal"])

		session.Close()
	}
}

func goTestFindCollectionName(t *testing.T) {
	logTestName()

	findCollectionName := func(entity interface{}) string {
		if _, ok := entity.(*User); ok {
			return "my users"
		}
		return ravendb.GetCollectionNameDefault(entity)
	}
	c := ravendb.NewDocumentConventions()
	c.FindCollectionName = findCollectionName
	name := c.GetCollectionName(&Employee{})
	assert.Equal(t, name, "Employees")

	name = c.GetCollectionName(&User{})
	assert.Equal(t, name, "my users")
}

// test that insertion order of bulk_docs (BatchOperation / BatchCommand)
func goTestBatchCommandOrder(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	// delete to trigger a code path that uses deferred commands
	// this is very sensitive to how code is structured: deleted
	// commands are gathered first in random order and put
	// commands are in insertion order
	nUsers := 10
	{
		session := openSessionMust(t, store)

		ids := []string{"users/5"}
		for i := 1; i <= nUsers; i++ {
			u := &User{
				Age: i,
			}
			u.setName(fmt.Sprintf("Name %d", i))
			id := fmt.Sprintf("users/%d", i)
			err = session.StoreWithID(u, id)
			assert.NoError(t, err)
			if i == 5 {
				err = session.Delete(u)
				assert.NoError(t, err)
			} else {
				ids = append(ids, id)
			}
		}
		commandsData, err := session.ForTestsSaveChangesGetCommands()
		assert.NoError(t, err)
		assert.Equal(t, len(commandsData), nUsers)
		for i, cmdData := range commandsData {
			var id string
			switch d := cmdData.(type) {
			case *ravendb.PutCommandDataWithJSON:
				id = d.ID
			case *ravendb.DeleteCommandData:
				id = d.ID
			}
			expID := ids[i]
			assert.Equal(t, expID, id)
			assert.Equal(t, expID, id)
		}

		session.Close()
	}
}

// test that we get a meaningful error for server exceptions sent as JSON response
// https://github.com/ravendb/ravendb-go-client/issues/147
func goTestInvalidIndexDefinition(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	restore := disableLogFailedRequests()
	defer restore()

	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	indexName := "Song/TextData"
	index := ravendb.NewIndexCreationTask(indexName)

	index.Map = `
from song in docs.Songs
select {
	SongData = new {
		song.Artist,
		song.Title,
		song.Tags,
		song.TrackId
	}
}
`
	index.Index("SongData", ravendb.FieldIndexingSearch)

	err = index.Execute(store, nil, "")
	assert.Error(t, err)
	_, ok := err.(*ravendb.IndexCompilationError)
	assert.True(t, ok)
}

// increasing code coverage of bulk_insert_operation.go
func goTestBulkInsertCoverage(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error
	store := driver.getDocumentStoreMust(t)

	var orphanedInsert *ravendb.BulkInsertOperation

	defer func() {
		restore := disableLogFailedRequests()
		store.Close()
		err = orphanedInsert.Close()
		assert.Error(t, err)
		restore()
	}()

	{

		bulkInsert := store.BulkInsert("")
		o := &FooBar{
			Name: "John Doe",
		}
		// trigger BulkInsertOperation.escapeID
		err = bulkInsert.StoreWithID(o, `FooBars/my-"-\id`, nil)
		assert.NoError(t, err)
		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		bulkInsert := store.BulkInsert("")
		o := &FooBar{
			Name: "John Doe",
		}
		err = bulkInsert.StoreWithID(o, ``, nil)
		assert.Error(t, err)
		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		bulkInsert := store.BulkInsert("")
		o := &FooBar{
			Name: "John Doe",
		}
		err = bulkInsert.StoreWithID(o, ``, nil)
		assert.Error(t, err)
		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		bulkInsert := store.BulkInsert("")
		o := &FooBar{
			Name: "John Doe",
		}
		// trigger a path in BulkInsertOperation.Store() that takes ID from metadata
		m := map[string]interface{}{
			ravendb.MetadataID: "FooBars/id-frommeta",
		}
		meta := ravendb.NewMetadataAsDictionaryWithMetadata(m)
		id, err := bulkInsert.Store(o, meta)
		assert.Equal(t, "FooBars/id-frommeta", id)
		assert.NoError(t, err)
		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		bulkInsert := store.BulkInsert("")
		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		orphanedInsert = store.BulkInsert("")
		o := &FooBar{
			Name: "John Doe",
		}
		_, err = orphanedInsert.Store(o, nil)
		assert.NoError(t, err)
	}

	{
		// try to trigger concurrency check
		bulkInsert := store.BulkInsert("")
		var wg sync.WaitGroup
		for i := 0; i < 5; i++ {
			wg.Add(1)
			go func() {
				o := &FooBar{
					Name: "John Doe",
				}
				_, _ = bulkInsert.Store(o, nil)
				wg.Done()
			}()
		}
		wg.Wait()

		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

	{
		// trigger operationID == -1 code path in Abort
		bulkInsert := store.BulkInsert("")
		err = bulkInsert.Abort()
		assert.NoError(t, err)
		err = bulkInsert.Close()
		assert.NoError(t, err)
	}

}

// increasing code coverage of raw_document_query.go
func goTestRawQueryCoverage(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error
	store := driver.getDocumentStoreMust(t)
	createNorthwindDatabase(t, driver, store)

	{
		session := openSessionMust(t, store)

		rawQuery := `from employees where FirstName == $p0`
		q := session.RawQuery(rawQuery)
		q = q.AddParameter("p0", "Anne")
		assert.NoError(t, q.Err())
		// adding the same parameter twice generates an error
		q = q.AddParameter("p0", "Anne")
		assert.Error(t, q.Err())
		q = q.AddParameter("p0", "Anne")
		assert.Error(t, q.Err())
		// trigger early error check
		q = q.UsingDefaultOperator(ravendb.QueryOperatorAnd)
		assert.Error(t, q.Err())

		// exercise error path in Any()
		_, _ = q.Any()

		session.Close()
	}

	{
		restore := disableLogFailedRequests()
		session := openSessionMust(t, store)
		rawQuery := `from employees where FirstName == $p0`
		q := session.RawQuery(rawQuery)
		// a no-op but exercises the code path
		q = q.UsingDefaultOperator(ravendb.QueryOperatorOr)

		var results []*northwind.Employee
		err = q.GetResults(&results)
		assert.Error(t, err)
		_, ok := err.(*ravendb.InvalidQueryError)
		assert.True(t, ok)

		session.Close()
		restore()
	}

	{
		session := openSessionMust(t, store)

		rawQuery := `from employees where FirstName == $p0`
		q := session.RawQuery(rawQuery)
		q = q.AddParameter("p0", "Anne")

		_, err = q.Any()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		rawQuery := `from employees where FirstName == $p0`
		q := session.RawQuery(rawQuery)
		q = q.AddParameter("p0", "Anne")
		q = q.WaitForNonStaleResultsWithTimeout(time.Second * 15)
		q = q.WaitForNonStaleResults()
		q = q.NoTracking()
		q = q.NoCaching()
		var stats *ravendb.QueryStatistics
		q = q.Statistics(&stats)

		nAfterQueryCalled := 0
		afterQueryExecuted := func(r *ravendb.QueryResult) {
			nAfterQueryCalled++
		}

		afterQueryExecutedIdx := q.AddAfterQueryExecutedListener(afterQueryExecuted)
		q = q.RemoveAfterQueryExecutedListener(afterQueryExecutedIdx)

		afterQueryExecutedIdx1 := q.AddAfterQueryExecutedListener(afterQueryExecuted)
		afterQueryExecutedIdx2 := q.AddAfterQueryExecutedListener(afterQueryExecuted)

		nBeforeQueryCalled := 0
		beforeQueryCalled := func(r *ravendb.IndexQuery) {
			nBeforeQueryCalled++
		}

		beforeQueryExecutedIdx := q.AddBeforeQueryExecutedListener(beforeQueryCalled)
		q = q.RemoveBeforeQueryExecutedListener(beforeQueryExecutedIdx)

		beforeQueryExecutedIdx1 := q.AddBeforeQueryExecutedListener(beforeQueryCalled)
		beforeQueryExecutedIdx2 := q.AddBeforeQueryExecutedListener(beforeQueryCalled)

		afterStreamExecuted := func(map[string]interface{}) {
			// no-op
		}
		afterStreamExecutedIdx := q.AddAfterStreamExecutedListener(afterStreamExecuted)

		var results []*northwind.Employee
		err = q.GetResults(&results)
		assert.NoError(t, err)

		q = q.RemoveAfterQueryExecutedListener(afterQueryExecutedIdx1)
		q = q.RemoveAfterQueryExecutedListener(afterQueryExecutedIdx2)

		q = q.RemoveBeforeQueryExecutedListener(beforeQueryExecutedIdx1)
		q = q.RemoveBeforeQueryExecutedListener(beforeQueryExecutedIdx2)

		q = q.RemoveAfterStreamExecutedListener(afterStreamExecutedIdx)
		assert.Equal(t, 2, nAfterQueryCalled)
		assert.Equal(t, 2, nBeforeQueryCalled)
		assert.NotNil(t, stats)

		session.Close()
	}
}

// increase code coverage in abstract_document_query.go etc.
func goTestQueryCoverage(t *testing.T, driver *RavenTestDriver) {
	logTestName()

	var err error
	store := driver.getDocumentStoreMust(t)
	createNorthwindDatabase(t, driver, store)

	{
		session := openSessionMust(t, store)
		q := session.QueryCollection("empoloyees")
		q = q.Distinct()
		_, err = q.Any()
		assert.NoError(t, err)
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		q := session.QueryCollection("empoloyees")
		// trigger error condition in distinct()
		q = q.Distinct()
		q = q.Distinct()
		assert.Error(t, q.Err())

		q = session.QueryCollection("empoloyees")
		q = q.Where("LastName", "asd", "me")
		assert.Error(t, q.Err())
		session.Close()
	}

	{
		session := openSessionMust(t, store)
		q := session.QueryCollection("empoloyees")
		q = q.Where("FirstName", "!=", "zzz")
		q = q.Where("FirstName", "<", "Zorro")
		q = q.Where("FirstName", "<=", "Zorro")
		q = q.Where("FirstName", ">", "Aha")
		q = q.Where("FirstName", ">=", "Aha")
		q = q.RandomOrderingWithSeed("")

		var results []*northwind.Employee
		err = q.GetResults(&results)
		assert.NoError(t, err)

		session.Close()
	}

}

func goTestLazyCoverage(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	{
		session := openSessionMust(t, store)
		for i := 1; i <= 2; i++ {
			company := &Company{
				ID: fmt.Sprintf("companies/%d", i),
			}
			err = session.StoreWithID(company, company.ID)
			assert.NoError(t, err)
		}

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		user := User5{
			Name: "Ayende",
		}
		err = session.Store(&user)
		assert.NoError(t, err)

		partner := User5{
			PartnerID: "user5s/1-A",
		}
		err = session.Store(&partner)
		assert.NoError(t, err)

		err = session.SaveChanges()
		assert.NoError(t, err)

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		fn1 := func() {
			// no-op
		}

		var company1Ref *Company

		query := session.Advanced().Lazily()
		// returns error on empty id
		lazy, err := query.LoadWithEval("", fn1, &company1Ref)
		assert.Error(t, err)
		assert.Nil(t, lazy)

		// returns error on empty ids
		lazy, err = query.LoadMulti(nil)
		assert.Error(t, err)
		assert.Nil(t, lazy)

		// returns error on empty ids
		lazy, err = query.LoadMultiWithEval(nil, fn1, nil)
		assert.Error(t, err)
		assert.Nil(t, lazy)

		var c *Company
		err = session.Load(&c, "companies/1")
		assert.NoError(t, err)
		assert.Equal(t, c.ID, "companies/1")

		// trigger o.delegate.IsLoaded(id) code path in LoadWithEval
		{
			query := session.Advanced().Lazily()
			lazy, err := query.LoadWithEval("companies/1", fn1, &company1Ref)
			assert.NoError(t, err)

			var c1 *Company
			err = lazy.GetValue(&c1)
			assert.NoError(t, err)
			assert.Equal(t, c.ID, "companies/1")
		}

		{
			session := openSessionMust(t, store)

			advanced := session.Advanced()
			_, err = advanced.Lazily().Load("user5s/2-A")
			assert.NoError(t, err)
			_, err = advanced.Lazily().Load("user5s/1-A")
			assert.NoError(t, err)

			_, err = advanced.Eagerly().ExecuteAllPendingLazyOperations()
			assert.NoError(t, err)

			oldCount := advanced.GetNumberOfRequests()

			resultLazy, err := advanced.Lazily().Include("PartnerId").Load("user5s/2-A")
			assert.NoError(t, err)
			var user *User
			err = resultLazy.GetValue(&user)
			assert.NoError(t, err)
			assert.NotNil(t, user)
			assert.Equal(t, user.ID, "user5s/2-A")

			newCount := advanced.GetNumberOfRequests()
			assert.Equal(t, newCount, oldCount)

			session.Close()
		}

		session.Close()
	}

	{
		session := openSessionMust(t, store)

		advanced := session.Advanced()

		{
			// empty id returns an error
			resultLazy, err := advanced.Lazily().Include("PartnerId").Load("")
			assert.Error(t, err)
			assert.Nil(t, resultLazy)
		}

		{
			// empty ids returns an error
			resultLazy, err := advanced.Lazily().Include("PartnerId").LoadMulti(nil)
			assert.Error(t, err)
			assert.Nil(t, resultLazy)

			resultLazy, err = advanced.Lazily().Include("PartnerId").LoadMulti([]string{})
			assert.Error(t, err)
			assert.Nil(t, resultLazy)
		}

		{
			resultLazy, err := advanced.Lazily().Include("PartnerId").LoadMulti([]string{"user5s/2-A", "user5s/1-A"})
			assert.NoError(t, err)
			err = resultLazy.GetValue(nil)
			assert.Error(t, err)
		}

		{
			resultLazy, err := advanced.Lazily().Include("PartnerId").LoadMulti([]string{"user5s/2-A", "user5s/1-A"})
			assert.NoError(t, err)
			results := map[string]*User5{}
			err = resultLazy.GetValue(results)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(results))
		}

		{
			resultLazy, err := advanced.Lazily().Include("PartnerId").LoadMulti([]string{"user5s/2-A", "user5s/1-A"})
			assert.NoError(t, err)
			results := map[string]*User{}
			// trying to get a mismatched type. has User5, we're trying to get User
			err = resultLazy.GetValue(results)
			assert.Error(t, err)
			assert.Equal(t, len(results), 0)
		}

		session.Close()
	}

}

func TestGo1(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	if false {
		goTestStoreMap(t, driver)
		goTest(t, driver)
		goTestGetLastModifiedForAndChanges(t, driver)
		goTestListeners(t, driver)
		goTestFindCollectionName(t)
		goTestBatchCommandOrder(t, driver)
		goTestInvalidIndexDefinition(t, driver)
		goTestBulkInsertCoverage(t, driver)
		goTestRawQueryCoverage(t, driver)
		goTestQueryCoverage(t, driver)
	}
	goTestLazyCoverage(t, driver)
}
