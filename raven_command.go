package ravendb

import (
	"io"
	"io/ioutil"
	"net/http"
)

var (
	_ RavenCommand = &RavenCommandBase{}
)

// RavenCommand defines interface for server commands
type RavenCommand interface {
	// those are meant to be over-written
	createRequest(node *ServerNode) (*http.Request, error)
	setResponse(response []byte, fromCache bool) error
	setResponseRaw(response *http.Response, body io.Reader) error

	// for all other functions, get access to underlying RavenCommandBase
	getBase() *RavenCommandBase
}

type RavenCommandBase struct {
	StatusCode           int
	ResponseType         RavenCommandResponseType
	CanCache             bool
	CanCacheAggressively bool

	// if true, can be cached
	IsReadRequest bool

	failedNodes map[*ServerNode]error
}

func NewRavenCommandBase() RavenCommandBase {
	res := RavenCommandBase{
		ResponseType:         RavenCommandResponseTypeObject,
		CanCache:             true,
		CanCacheAggressively: true,
	}
	return res
}

func (c *RavenCommandBase) getBase() *RavenCommandBase {
	return c
}

func (c *RavenCommandBase) setResponse(response []byte, fromCache bool) error {
	if c.ResponseType == RavenCommandResponseTypeEmpty || c.ResponseType == RavenCommandResponseTypeRaw {
		return throwInvalidResponse()
	}

	return newUnsupportedOperationError(c.ResponseType + " command must override the SetResponse method which expects response with the following type: " + c.ResponseType)
}

func (c *RavenCommandBase) setResponseRaw(response *http.Response, stream io.Reader) error {
	panicIf(true, "When "+c.ResponseType+" is set to Raw then please override this method to handle the response. ")
	return nil
}

func (c *RavenCommandBase) createRequest(node *ServerNode) (*http.Request, error) {
	panicIf(true, "createRequest must be over-written by all types")
	return nil, nil
}

func throwInvalidResponse() error {
	return newIllegalStateError("Invalid response")
}

func (c *RavenCommandBase) Send(client *http.Client, req *http.Request) (*http.Response, error) {
	rsp, err := client.Do(req)
	return rsp, err
}

func (c *RavenCommandBase) GetFailedNodes() map[*ServerNode]error {
	return c.failedNodes
}

func (c *RavenCommandBase) SetFailedNodes(failedNodes map[*ServerNode]error) {
	c.failedNodes = failedNodes
}

func (c *RavenCommandBase) urlEncode(value string) string {
	return urlEncode(value)
}

func ensureIsNotNullOrString(value string, name string) error {
	if value == "" {
		return newIllegalArgumentError("%s cannot be null or empty", name)
	}
	return nil
}

func (c *RavenCommandBase) IsFailedWithNode(node *ServerNode) bool {
	if c.failedNodes == nil {
		return false
	}
	_, ok := c.failedNodes[node]
	return ok
}

// Note: in Java this is part of RavenCommand and can be virtual
// That's imposssible in Go, so we replace with stand-alone function
func processCommandResponse(cmd RavenCommand, cache *HttpCache, response *http.Response, url string) (responseDisposeHandling, error) {
	// In Java this is overridden in HeadDocumentCommand, so hack it this way
	if cmdHead, ok := cmd.(*HeadDocumentCommand); ok {
		return cmdHead.ProcessResponse(cache, response, url)
	}

	if cmdHead, ok := cmd.(*HeadAttachmentCommand); ok {
		return cmdHead.processResponse(cache, response, url)
	}

	if cmdGet, ok := cmd.(*GetAttachmentCommand); ok {
		return cmdGet.processResponse(cache, response, url)
	}

	if cmdQuery, ok := cmd.(*QueryStreamCommand); ok {
		return cmdQuery.processResponse(cache, response, url)
	}

	if cmdStream, ok := cmd.(*StreamCommand); ok {
		return cmdStream.processResponse(cache, response, url)
	}

	//fmt.Printf("processCommandResponse of %T\n", cmd)
	c := cmd.getBase()

	if response.Body == nil {
		return responseDisposeHandlingAutomatic, nil
	}

	statusCode := response.StatusCode
	if c.ResponseType == RavenCommandResponseTypeEmpty || statusCode == http.StatusNoContent {
		return responseDisposeHandlingAutomatic, nil
	}

	if c.ResponseType == RavenCommandResponseTypeObject {
		contentLength := response.ContentLength
		if contentLength == 0 {
			return responseDisposeHandlingAutomatic, nil
		}

		// we intentionally don't dispose the reader here, we'll be using it
		// in the command, any associated memory will be released on context reset
		js, err := ioutil.ReadAll(response.Body)
		if err != nil {
			return responseDisposeHandlingAutomatic, err
		}

		if cache != nil {
			//fmt.Printf("processCommandResponse: caching response for %s\n", url)
			c.CacheResponse(cache, url, response, js)
		}
		err = cmd.setResponse(js, false)
		return responseDisposeHandlingAutomatic, err
	}

	err := cmd.setResponseRaw(response, response.Body)
	return responseDisposeHandlingAutomatic, err
}

func (c *RavenCommandBase) CacheResponse(cache *HttpCache, url string, response *http.Response, responseJson []byte) {
	if !c.CanCache {
		//fmt.Printf("CacheResponse: url: %s, !c.CanCache\n", url)
		return
	}

	changeVector := gttpExtensionsGetEtagHeader(response)
	if changeVector == nil {
		//fmt.Printf("CacheResponse: url: %s, not caching because changeVector==nil\n", url)
		return
	}

	cache.set(url, changeVector, responseJson)
}

func (c *RavenCommandBase) AddChangeVectorIfNotNull(changeVector *string, request *http.Request) {
	if changeVector != nil {
		request.Header.Add("If-Match", `"`+*changeVector+`"`)
	}
}

func (c *RavenCommandBase) OnResponseFailure(response *http.Response) {
	// TODO: it looks like it's meant to be virtual but there are no
	// over-rides in Java code
}

// Note: hackish solution due to lack of generics
// For commands whose result is OperationIDResult, return it
// When new command returning OperationIDResult are added, we must extend it
func getCommandOperationIDResult(cmd RavenCommand) *OperationIDResult {
	switch c := cmd.(type) {
	case *CompactDatabaseCommand:
		return c.Result
	case *PatchByQueryCommand:
		return c.Result
	case *DeleteByIndexCommand:
		return c.Result
	}

	panicIf(true, "called on a command %T that doesn't return OperationIDResult", cmd)
	return nil
}
