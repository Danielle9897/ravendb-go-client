package ravendb

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

type IDocumentStore = DocumentStore

// DocumentStore represents a database
type DocumentStore struct {
	// from DocumentStoreBase
	onBeforeStore      []func(interface{}, *BeforeStoreEventArgs)
	onAfterSaveChanges []func(interface{}, *AfterSaveChangesEventArgs)

	onBeforeDelete   []func(interface{}, *BeforeDeleteEventArgs)
	onBeforeQuery    []func(interface{}, *BeforeQueryEventArgs)
	onSessionCreated []func(interface{}, *SessionCreatedEventArgs)

	disposed    bool
	conventions *DocumentConventions
	urls        []string // urls for HTTP endopoints of server nodes
	initialized bool
	certificate *KeyStore
	database    string // name of the database

	// maps database name to databaseChanges. Must be protected with mutex
	databaseChanges map[string]*databaseChanges

	// Note: access must be protected with mu
	// Lazy.Value is **EvictItemsFromCacheBasedOnChanges
	aggressiveCacheChanges map[string]*Lazy

	// maps database name to its RequestsExecutor
	// access must be protected with mu
	// TODO: in Java is ConcurrentMap<String, RequestExecutor> requestExecutors
	// so must protect access with mutex and use case-insensitive lookup
	requestsExecutors map[string]*RequestExecutor

	multiDbHiLo                  *MultiDatabaseHiLoIDGenerator
	maintenanceOperationExecutor *MaintenanceOperationExecutor
	operationExecutor            *OperationExecutor
	identifier                   string
	aggressiveCachingUsed        bool

	afterClose  []func(*DocumentStore)
	beforeClose []func(*DocumentStore)

	mu sync.Mutex
}

// from DocumentStoreBase
func (s *DocumentStore) GetConventions() *DocumentConventions {
	if s.conventions == nil {
		s.conventions = NewDocumentConventions()
	}
	return s.conventions
}

func (s *DocumentStore) SetConventions(conventions *DocumentConventions) {
	s.conventions = conventions
}

func (s *DocumentStore) GetUrls() []string {
	return s.urls
}

func (s *DocumentStore) SetUrls(value []string) {
	panicIf(len(value) == 0, "value is empty")
	for i, s := range value {
		value[i] = strings.TrimSuffix(s, "/")
	}
	s.urls = value
}

func (s *DocumentStore) ensureNotClosed() {
	// TODO: implement me
}

func (s *DocumentStore) AddBeforeStoreListener(handler func(interface{}, *BeforeStoreEventArgs)) int {
	s.onBeforeStore = append(s.onBeforeStore, handler)
	return len(s.onBeforeStore) - 1

}
func (s *DocumentStore) RemoveBeforeStoreListener(handlerIdx int) {
	s.onBeforeStore[handlerIdx] = nil
}

func (s *DocumentStore) AddAfterSaveChangesListener(handler func(interface{}, *AfterSaveChangesEventArgs)) int {
	s.onAfterSaveChanges = append(s.onAfterSaveChanges, handler)
	return len(s.onAfterSaveChanges) - 1
}

func (s *DocumentStore) RemoveAfterSaveChangesListener(handlerIdx int) {
	s.onAfterSaveChanges[handlerIdx] = nil
}

func (s *DocumentStore) AddBeforeDeleteListener(handler func(interface{}, *BeforeDeleteEventArgs)) int {
	s.onBeforeDelete = append(s.onBeforeDelete, handler)
	return len(s.onBeforeDelete) - 1
}

func (s *DocumentStore) RemoveBeforeDeleteListener(handlerIdx int) {
	s.onBeforeDelete[handlerIdx] = nil
}

func (s *DocumentStore) AddBeforeQueryListener(handler func(interface{}, *BeforeQueryEventArgs)) int {
	s.onBeforeQuery = append(s.onBeforeQuery, handler)
	return len(s.onBeforeQuery) - 1
}

func (s *DocumentStore) RemoveBeforeQueryListener(handlerIdx int) {
	s.onBeforeQuery[handlerIdx] = nil
}

func (s *DocumentStore) RegisterEvents(session *InMemoryDocumentSessionOperations) {
	// TODO: unregister those events?
	for _, handler := range s.onBeforeStore {
		if handler != nil {
			session.AddBeforeStoreListener(handler)
		}
	}

	for _, handler := range s.onAfterSaveChanges {
		if handler != nil {
			session.AddAfterSaveChangesListener(handler)
		}
	}

	for _, handler := range s.onBeforeDelete {
		if handler != nil {
			session.AddBeforeDeleteListener(handler)
		}
	}

	for _, handler := range s.onBeforeQuery {
		if handler != nil {
			session.AddBeforeQueryListener(handler)
		}
	}
}

func (s *DocumentStore) afterSessionCreated(session *InMemoryDocumentSessionOperations) {
	for _, handler := range s.onSessionCreated {
		if handler != nil {
			args := &SessionCreatedEventArgs{
				Session: session,
			}
			handler(s, args)
		}
	}
}

func (s *DocumentStore) assertInitialized() {
	panicIf(!s.initialized, "DocumentStore must be initialized")
}

func (s *DocumentStore) GetDatabase() string {
	return s.database
}

func (s *DocumentStore) SetDatabase(database string) {
	panicIf(s.initialized, "is already initialized")
	s.database = database
}

func (s *DocumentStore) GetCertificate() *KeyStore {
	return s.certificate
}

func (s *DocumentStore) SetCertificate(certificate *KeyStore) {
	panicIf(s.initialized, "is already initialized")
	s.certificate = certificate
}

func (s *DocumentStore) AggressivelyCache() {
	s.AggressivelyCacheWithDatabase("")
}

func (s *DocumentStore) AggressivelyCacheWithDatabase(database string) {
	s.AggressivelyCacheForDatabase(time.Hour*24, database)
}

// NewDocumentStore creates a DocumentStore
func NewDocumentStore() *DocumentStore {
	s := &DocumentStore{
		requestsExecutors:      map[string]*RequestExecutor{},
		conventions:            NewDocumentConventions(),
		databaseChanges:        map[string]*databaseChanges{},
		aggressiveCacheChanges: map[string]*Lazy{},
	}
	return s
}

func NewDocumentStoreWithURLAndDatabase(url string, database string) *DocumentStore {
	res := NewDocumentStore()
	res.SetUrls([]string{url})
	res.SetDatabase(database)
	return res
}

func NewDocumentStoreWithURLsAndDatabase(urls []string, database string) *DocumentStore {
	res := NewDocumentStore()
	res.SetUrls(urls)
	res.SetDatabase(database)
	return res
}

func (s *DocumentStore) GetIdentifier() string {
	if s.identifier != "" {
		return s.identifier
	}

	if len(s.urls) == 0 {
		return ""
	}

	if s.database != "" {
		return strings.Join(s.urls, ",") + " (DB: " + s.database + ")"
	}

	return strings.Join(s.urls, ",")
}

func (s *DocumentStore) SetIdentifier(identifier string) {
	s.identifier = identifier
}

// Close closes the Store
func (s *DocumentStore) Close() {
	if s.disposed {
		return
	}

	for _, fn := range s.beforeClose {
		fn(s)
	}
	s.beforeClose = nil

	for _, value := range s.aggressiveCacheChanges {
		if !value.IsValueCreated() {
			continue
		}

		err := value.GetValue()
		if err != nil {
			v := value.Value.(**EvictItemsFromCacheBasedOnChanges)
			if v != nil {
				(*v).Close()
			}
		}
	}

	for _, changes := range s.databaseChanges {
		changes.Close()
	}

	if s.multiDbHiLo != nil {
		s.multiDbHiLo.ReturnUnusedRange()
	}

	s.disposed = true

	for _, fn := range s.afterClose {
		fn(s)
	}
	s.afterClose = nil

	for _, re := range s.requestsExecutors {
		re.Close()
	}
}

// OpenSession opens a new session to document Store.
func (s *DocumentStore) OpenSession() (*DocumentSession, error) {
	return s.OpenSessionWithOptions(&SessionOptions{})
}

func (s *DocumentStore) OpenSessionWithDatabase(database string) (*DocumentSession, error) {
	sessionOptions := &SessionOptions{
		Database: database,
	}
	return s.OpenSessionWithOptions(sessionOptions)
}

func (s *DocumentStore) OpenSessionWithOptions(options *SessionOptions) (*DocumentSession, error) {
	s.assertInitialized()
	s.ensureNotClosed()

	sessionID := NewUUID().String()
	databaseName := firstNonEmptyString(options.Database, s.GetDatabase())
	requestExecutor := options.RequestExecutor
	if requestExecutor == nil {
		requestExecutor = s.GetRequestExecutor(databaseName)
	}
	session := NewDocumentSession(databaseName, s, sessionID, requestExecutor)
	s.RegisterEvents(session.InMemoryDocumentSessionOperations)
	s.afterSessionCreated(session.InMemoryDocumentSessionOperations)
	return session, nil
}

func (s *DocumentStore) ExecuteIndex(task *AbstractIndexCreationTask) error {
	return s.ExecuteIndexWithDatabase(task, "")
}

func (s *DocumentStore) ExecuteIndexWithDatabase(task *AbstractIndexCreationTask, database string) error {
	s.assertInitialized()
	return task.Execute2(s, s.conventions, database)
}

func (s *DocumentStore) ExecuteIndexes(tasks []*AbstractIndexCreationTask) error {
	return s.ExecuteIndexesWithDatabase(tasks, "")
}

func (s *DocumentStore) ExecuteIndexesWithDatabase(tasks []*AbstractIndexCreationTask, database string) error {
	s.assertInitialized()
	indexesToAdd := indexCreationCreateIndexesToAdd(tasks, s.conventions)

	op := NewPutIndexesOperation(indexesToAdd...)
	if database == "" {
		database = s.GetDatabase()
	}
	return s.Maintenance().ForDatabase(database).Send(op)
}

// GetRequestExecutorWithDatabase gets a request executor for a given database
// database is optional
func (s *DocumentStore) GetRequestExecutor(database string) *RequestExecutor {
	s.assertInitialized()
	if database == "" {
		database = s.GetDatabase()
	}
	database = strings.ToLower(database)

	s.mu.Lock()
	executor, ok := s.requestsExecutors[database]
	s.mu.Unlock()

	if ok {
		return executor
	}

	if !s.GetConventions().IsDisableTopologyUpdates() {
		executor = RequestExecutorCreate(s.GetUrls(), s.GetDatabase(), s.GetCertificate(), s.GetConventions())
	} else {
		executor = RequestExecutorCreateForSingleNodeWithConfigurationUpdates(s.GetUrls()[0], s.GetDatabase(), s.GetCertificate(), s.GetConventions())
	}

	s.mu.Lock()
	s.requestsExecutors[database] = executor
	s.mu.Unlock()

	return executor
}

// Initialize initializes document Store,
// Must be called before executing any operation.
func (s *DocumentStore) Initialize() error {
	if s.initialized {
		return nil
	}
	err := s.assertValidConfiguration()
	if err != nil {
		return err
	}

	conventions := s.conventions
	if conventions.GetDocumentIDGenerator() == nil {
		generator := NewMultiDatabaseHiLoIDGenerator(s, s.GetConventions())
		s.multiDbHiLo = generator
		genID := func(dbName string, entity interface{}) string {
			return generator.GenerateDocumentID(dbName, entity)
		}
		conventions.SetDocumentIDGenerator(genID)
	}
	s.initialized = true
	return nil
}

func (s *DocumentStore) assertValidConfiguration() error {
	if len(s.urls) == 0 {
		return fmt.Errorf("Must provide urls to NewDocumentStore")
	}
	return nil
}

type RestoreCaching struct {
	re  *RequestExecutor
	old *AggressiveCacheOptions
}

func (r *RestoreCaching) Close() error {
	r.re.aggressiveCaching = r.old
	return nil
}

func (s *DocumentStore) DisableAggressiveCaching() *RestoreCaching {
	return s.DisableAggressiveCachingWithDatabase("")
}

func (s *DocumentStore) DisableAggressiveCachingWithDatabase(databaseName string) *RestoreCaching {
	if databaseName == "" {
		databaseName = s.GetDatabase()
	}

	re := s.GetRequestExecutor(databaseName)
	old := re.aggressiveCaching
	re.aggressiveCaching = nil
	res := &RestoreCaching{
		re:  re,
		old: old,
	}
	return res
}

func (s *DocumentStore) Changes() *databaseChanges {
	return s.ChangesWithDatabaseName("")
}

func (s *DocumentStore) ChangesWithDatabaseName(database string) *databaseChanges {
	s.assertInitialized()

	if database == "" {
		database = s.GetDatabase()
	}

	s.mu.Lock()
	changes, ok := s.databaseChanges[database]
	s.mu.Unlock()

	if !ok {
		changes = s.createDatabaseChanges(database)

		s.mu.Lock()
		s.databaseChanges[database] = changes
		s.mu.Unlock()

	}
	return changes
}

func (s *DocumentStore) createDatabaseChanges(database string) *databaseChanges {
	onDispose := func() {
		s.mu.Lock()
		delete(s.databaseChanges, database)
		s.mu.Unlock()
	}
	re := s.GetRequestExecutor(database)
	return newDatabaseChanges(re, database, onDispose)
}

func (s *DocumentStore) GetLastDatabaseChangesStateError() error {
	return s.GetLastDatabaseChangesStateErrorWithDatabaseName("")
}

func (s *DocumentStore) GetLastDatabaseChangesStateErrorWithDatabaseName(database string) error {
	if database == "" {
		database = s.GetDatabase()
	}

	s.mu.Lock()
	databaseChanges, ok := s.databaseChanges[database]
	s.mu.Unlock()

	if !ok {
		return nil
	}
	ch := databaseChanges
	return ch.getLastConnectionStateError()
}

func (s *DocumentStore) AggressivelyCacheFor(cacheDuration time.Duration) io.Closer {
	return s.AggressivelyCacheForDatabase(cacheDuration, "")
}

type aggressiveCachingRestorer struct {
	re  *RequestExecutor
	old *AggressiveCacheOptions
}

func (r *aggressiveCachingRestorer) Close() error {
	r.re.aggressiveCaching = r.old
	return nil
}

func (s *DocumentStore) AggressivelyCacheForDatabase(cacheDuration time.Duration, database string) io.Closer {
	if database == "" {
		database = s.GetDatabase()
	}
	panicIf(database == "", "must have database") // TODO: maybe return error
	if !s.aggressiveCachingUsed {
		s.listenToChangesAndUpdateTheCache(database)
	}

	re := s.GetRequestExecutor(database)
	old := re.aggressiveCaching

	opts := &AggressiveCacheOptions{
		Duration: cacheDuration,
	}
	re.aggressiveCaching = opts
	restorer := &aggressiveCachingRestorer{
		re:  re,
		old: old,
	}
	return restorer
}

func (s *DocumentStore) listenToChangesAndUpdateTheCache(database string) {
	// this is intentionally racy, most cases, we'll already
	// have this set once, so we won't need to do it again
	s.aggressiveCachingUsed = true

	s.mu.Lock()
	lazy := s.aggressiveCacheChanges[database]
	s.mu.Unlock()

	if lazy == nil {
		valueFactory := func(result interface{}) error {
			res := NewEvictItemsFromCacheBasedOnChanges(s, database)
			resultPtr := result.(**EvictItemsFromCacheBasedOnChanges)
			*resultPtr = res
			return nil
		}
		var results *EvictItemsFromCacheBasedOnChanges
		lazy = NewLazy(&results, valueFactory)

		s.mu.Lock()
		s.aggressiveCacheChanges[database] = lazy
		s.mu.Unlock()
	}

	lazy.GetValue() // force evaluation
}

func (s *DocumentStore) AddBeforeCloseListener(fn func(*DocumentStore)) int {
	s.beforeClose = append(s.beforeClose, fn)
	return len(s.beforeClose) - 1
}

func (s *DocumentStore) RemoveBeforeCloseListener(idx int) {
	s.beforeClose[idx] = nil
}

func (s *DocumentStore) AddAfterCloseListener(fn func(*DocumentStore)) int {
	s.afterClose = append(s.afterClose, fn)
	return len(s.afterClose) - 1
}

func (s *DocumentStore) RemoveAfterCloseListener(idx int) {
	s.afterClose[idx] = nil
}

func (s *DocumentStore) Maintenance() *MaintenanceOperationExecutor {
	s.assertInitialized()

	if s.maintenanceOperationExecutor == nil {
		s.maintenanceOperationExecutor = NewMaintenanceOperationExecutor(s)
	}

	return s.maintenanceOperationExecutor
}

func (s *DocumentStore) Operations() *OperationExecutor {
	if s.operationExecutor == nil {
		s.operationExecutor = NewOperationExecutor(s)
	}

	return s.operationExecutor
}

func (s *DocumentStore) BulkInsert() *BulkInsertOperation {
	return s.BulkInsertWithDatabase("")
}

func (s *DocumentStore) BulkInsertWithDatabase(database string) *BulkInsertOperation {
	if database == "" {
		database = s.GetDatabase()
	}
	return NewBulkInsertOperation(database, s)
}
