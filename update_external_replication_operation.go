package ravendb

import (
	"encoding/json"
	"net/http"
)

var (
	_ IMaintenanceOperation = &UpdateExternalReplicationOperation{}
)

type UpdateExternalReplicationOperation struct {
	_newWatcher *ExternalReplication

	Command *UpdateExternalReplicationCommand
}

func NewUpdateExternalReplicationOperation(newWatcher *ExternalReplication) *UpdateExternalReplicationOperation {
	return &UpdateExternalReplicationOperation{
		_newWatcher: newWatcher,
	}
}

func (o *UpdateExternalReplicationOperation) getCommand(conventions *DocumentConventions) RavenCommand {
	o.Command = NewUpdateExternalReplicationCommand(o._newWatcher)
	return o.Command
}

var _ RavenCommand = &UpdateExternalReplicationCommand{}

type UpdateExternalReplicationCommand struct {
	*RavenCommandBase

	_newWatcher *ExternalReplication

	Result *ModifyOngoingTaskResult
}

func NewUpdateExternalReplicationCommand(newWatcher *ExternalReplication) *UpdateExternalReplicationCommand {
	cmd := &UpdateExternalReplicationCommand{
		RavenCommandBase: NewRavenCommandBase(),

		_newWatcher: newWatcher,
	}
	return cmd
}

func (c *UpdateExternalReplicationCommand) createRequest(node *ServerNode) (*http.Request, error) {
	url := node.getUrl() + "/databases/" + node.getDatabase() + "/admin/tasks/external-replication"

	m := map[string]interface{}{
		"Watcher": c._newWatcher,
	}
	d, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return NewHttpPost(url, d)
}

func (c *UpdateExternalReplicationCommand) setResponse(response []byte, fromCache bool) error {
	if len(response) == 0 {
		return throwInvalidResponse()
	}

	var res ModifyOngoingTaskResult
	err := json.Unmarshal(response, &res)
	if err != nil {
		return err
	}
	c.Result = &res
	return nil
}
