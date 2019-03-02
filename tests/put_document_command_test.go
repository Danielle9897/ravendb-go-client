package tests

import (
	"testing"

	"github.com/ravendb/ravendb-go-client"
	"github.com/stretchr/testify/assert"
)

func putDocumentCommandCanPutDocumentUsingCommand(t *testing.T, driver *RavenTestDriver) {
	var err error
	store := driver.getDocumentStoreMust(t)
	defer store.Close()

	user := &User{}
	user.setName("Marcin")
	user.Age = 30

	node, err := entityToDocument(user)
	assert.NoError(t, err)

	command := ravendb.NewPutDocumentCommand("users/1", nil, node)
	err = store.GetRequestExecutor("").ExecuteCommand(command, nil)
	assert.NoError(t, err)

	result := command.Result
	assert.Equal(t, "users/1", result.ID)

	assert.NotNil(t, result.ChangeVector)

	{
		session := openSessionMust(t, store)
		var loadedUser *User
		err = session.Load(&loadedUser, "users/1")
		assert.NoError(t, err)
		assert.Equal(t, "Marcin", *loadedUser.Name)
		session.Close()
	}
}

func TestPutDocumentCommand(t *testing.T) {
	driver := createTestDriver(t)
	destroy := func() { destroyDriver(t, driver) }
	defer recoverTest(t, destroy)

	putDocumentCommandCanPutDocumentUsingCommand(t, driver)
}
