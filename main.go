package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
)

// Entity

type Id = int
type Name = string
type Content = string

type Note struct {
	id      Id
	name    Name
	content Content
}

type NoteList []Note

// Storage
type Storage interface {
	ReadAll() NoteList
	Read(Id) Note
	Create(Name, Content) Note
	Update(Id, Name, Content) Note
	Delete(Id) Note
}

var id Id = 0
var noteMap map[Id]Note = map[Id]Note{}

// InMemoryStorage saves data in memory during the programme execution
// there is no persistance
type InMemoryStorage struct{}

func (s InMemoryStorage) Read(id Id) Note {
	return noteMap[id]
}

func (s InMemoryStorage) ReadAll() NoteList {
	notes := NoteList{}
	for _, v := range noteMap {
		notes = append(notes, v)
	}
	return notes
}

func (s InMemoryStorage) Create(name Name, content Content) Note {
	newId := id + 1
	id = newId
	newNote := Note{
		id:      newId,
		name:    name,
		content: content,
	}
	noteMap[newId] = newNote
	return newNote
}

func (s InMemoryStorage) Update(id Id, name Name, content Content) Note {
	note := noteMap[id]
	if name != "" {
		note.name = name
	}
	if content != "" {
		note.content = content
	}
	noteMap[id] = note
	return note
}

func (s InMemoryStorage) Delete(id Id) Note {
	note := noteMap[id]
	delete(noteMap, id)
	return note
}

// Json Storage

// Command
type Command[Message, Result any] interface {
	execute(Message) Result
}

// ReadAll usecase
type ReadAllMessage struct{}

type ReadAllResult struct {
	notes []Note
}

type ReadAllCommand struct {
	storage Storage
}

func (u ReadAllCommand) execute(i ReadAllMessage) ReadAllResult {
	notes := u.storage.ReadAll()
	return ReadAllResult{
		notes: notes,
	}
}

// Read usecase
type ReadCommand struct {
	storage Storage
}
type ReadMessage struct {
	id Id
}
type ReadResult struct {
	note Note
}

func (u ReadCommand) execute(i ReadMessage) ReadResult {
	note := u.storage.Read(i.id)
	return ReadResult{
		note: note,
	}
}

// Create usecase
type CreateCommand struct {
	storage Storage
}
type CreateMessage struct {
	name    Name
	content Content
}
type CreateResult struct {
	note Note
}

func (u CreateCommand) execute(i CreateMessage) CreateResult {
	note := u.storage.Create(i.name, i.content)
	return CreateResult{
		note: note,
	}
}

// Update usecase
type UpdateCommand struct {
	storage Storage
}
type UpdateMessage struct {
	id      Id
	name    Name
	content Content
}
type UpdateResult struct {
	note Note
}

func (u UpdateCommand) execute(i UpdateMessage) UpdateResult {
	note := u.storage.Update(i.id, i.name, i.content)
	return UpdateResult{
		note: note,
	}
}

// Delete Command
type DeleteCommand struct {
	storage Storage
}
type DeleteMessage struct {
	id Id
}
type DeleteResult struct {
	note Note
}

func (u DeleteCommand) execute(i DeleteMessage) DeleteResult {
	note := u.storage.Delete(i.id)
	return DeleteResult{
		note: note,
	}
}

type Usecase struct {
	read    ReadCommand
	readAll ReadAllCommand
	create  CreateCommand
	update  UpdateCommand
	delete  DeleteCommand
}

// Inversion of control happens here
// Usecase only know the storage interface which could have
// many implementations
func newUsecase(storage Storage) Usecase {
	return Usecase{
		ReadCommand{storage},
		ReadAllCommand{storage},
		CreateCommand{storage},
		UpdateCommand{storage},
		DeleteCommand{storage},
	}
}

// Input Parser
type Parser[I any] interface {
	fromRepl(I) Parser[I]
	fromHttp(I) Parser[I]
}

type ReadAllParser struct{}

func (c ReadAllParser) fromHttp(r *http.Request) ReadAllMessage {
	return ReadAllMessage{}
}

func (c ReadAllParser) fromRepl(s []string) ReadAllMessage {
	return ReadAllMessage{}
}

type ReadParser struct{}

func (c ReadParser) fromHttp(r *http.Request) ReadMessage {
	return ReadMessage{}
}

func (c ReadParser) fromRepl(s []string) ReadMessage {
	id := s[1]
	number, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	return ReadMessage{
		id: number,
	}
}

type CreateParser struct{}

func (c CreateParser) fromHttp(r *http.Request) CreateMessage {
	return CreateMessage{}
}

func (c CreateParser) fromRepl(s []string) CreateMessage {
	name := s[1]
	content := s[2]
	return CreateMessage{
		name:    name,
		content: content,
	}
}

type UpdateParser struct{}

func (c UpdateParser) fromHttp(r *http.Request) UpdateMessage {
	return UpdateMessage{}
}

func (c UpdateParser) fromRepl(s []string) UpdateMessage {
	id := s[1]
	number, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	name := s[2]
	content := s[3]
	return UpdateMessage{
		id:      number,
		name:    name,
		content: content,
	}
}

type DeleteParser struct{}

func (c DeleteParser) fromHttp(r *http.Request) DeleteMessage {
	return DeleteMessage{}
}

func (c DeleteParser) fromRepl(s []string) DeleteMessage {
	return DeleteMessage{}
}

type ParserHandler struct {
	readParser    ReadParser
	readAllParser ReadAllParser
	createParser  CreateParser
	updateParser  UpdateParser
	deleteParser  DeleteParser
}

// Presenter
type Presenter[T any] interface {
	present(any, T)
}

type JsonPresenter struct{}

func (p JsonPresenter) present(o any, w http.ResponseWriter) {
	json.NewEncoder(w).Encode(o)
}

type ReplPresenter struct{}

func (p ReplPresenter) present(o any, _ any) {
	fmt.Println(o)
}

// Application
type Application interface {
	run()
}

// Repl Application
type ReplApplication struct {
	parser    ParserHandler
	usecase   Usecase
	presenter ReplPresenter
}

func (app ReplApplication) handleReadAll(input []string) {
	message := app.parser.readAllParser.fromRepl(input)
	result := app.usecase.readAll.execute(message)
	app.presenter.present(result, nil)
}

func (app ReplApplication) handleRead(input []string) {
	message := app.parser.readParser.fromRepl(input)
	result := app.usecase.read.execute(message)
	app.presenter.present(result, nil)
}

func (app ReplApplication) handleCreate(input []string) {
	message := app.parser.createParser.fromRepl(input)
	result := app.usecase.create.execute(message)
	app.presenter.present(result, nil)
}

func (app ReplApplication) handleUpdate(input []string) {
	message := app.parser.updateParser.fromRepl(input)
	result := app.usecase.update.execute(message)
	app.presenter.present(result, nil)
}

func (app ReplApplication) handleDelete(input []string) {
	message := app.parser.deleteParser.fromRepl(input)
	result := app.usecase.delete.execute(message)
	app.presenter.present(result, nil)
}

func (ReplApplication) shouldExit(input string) bool {
	return strings.TrimSpace(input) == "exit"
}

func (app ReplApplication) run() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("REPL > ")
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(err)
		}
		if app.shouldExit(input) {
			break
		}
		args := strings.Split(input, ";")
		for i := range args {
			args[i] = strings.TrimSpace(args[i])
		}
		switch args[0] {
		case "CREATE":
			app.handleCreate(args)
		case "READ":
			app.handleRead(args)
		case "READALL":
			app.handleReadAll(args)
		case "UPDATE":
			app.handleUpdate(args)
		case "DELETE":
			app.handleDelete(args)
		default:
			panic("Unknown command")
		}
	}
}

// HttpApplication
type HttpApplication struct {
	parser    ParserHandler
	usecase   Usecase
	presenter JsonPresenter
}

func (app HttpApplication) handleGet(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		message := app.parser.readAllParser.fromHttp(r)
		result := app.usecase.readAll.execute(message)
		app.presenter.present(result, w)
		return
	}
	_, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}
	message := app.parser.readParser.fromHttp(r)
	result := app.usecase.read.execute(message)
	app.presenter.present(result, w)
}

func (app HttpApplication) handlePost(w http.ResponseWriter, r *http.Request) {
	message := app.parser.createParser.fromHttp(r)
	result := app.usecase.create.execute(message)
	app.presenter.present(result, w)
}

func (app HttpApplication) handlePut(w http.ResponseWriter, r *http.Request) {
	message := app.parser.updateParser.fromHttp(r)
	result := app.usecase.update.execute(message)
	app.presenter.present(result, w)
}

func (app HttpApplication) handleDelete(w http.ResponseWriter, r *http.Request) {
	message := app.parser.deleteParser.fromHttp(r)
	result := app.usecase.delete.execute(message)
	app.presenter.present(result, w)
}

func (app HttpApplication) run() {
	http.HandleFunc("/notes/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			app.handleGet(w, r)
		case "POST":
			app.handlePost(w, r)
		case "PUT":
			app.handlePut(w, r)
		case "DELETE":
			app.handleDelete(w, r)
		default:
			panic("Uknown method")
		}
	})
	http.ListenAndServe("127.0.0.1:80", nil)
}

type AppMode string

const (
	HTTP AppMode = "HTTP"
	REPL AppMode = "REPL"
)

func newApplication(mode AppMode) Application {
	var app Application
	storage := InMemoryStorage{}
	switch mode {
	case REPL:
		app = ReplApplication{
			usecase: newUsecase(storage),
		}
	case HTTP:
		app = HttpApplication{
			usecase: newUsecase(storage),
		}
	default:
		panic("Unknown application mode")
	}
	return app
}

func main() {
	newApplication(REPL).run()
}
