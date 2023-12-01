package main

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type User struct {
	Id       uuid.UUID
	Projects []Project
}

type Project struct {
	Id                uuid.UUID
	Name              string
	UncategoriedTasks []Task
	PrimaryColor      string
}

type Task struct {
	Id          uuid.UUID
	Name        string
	Description string
	Finished    bool
	StartDate   *time.Time
}

func (t *Task) SetFinished(e bool) {
	t.Finished = e
}

var users []User

var exampleProjects []Project

type Unauthorized struct{}

func (err Unauthorized) Error() string {
	return "Unauthorized access"
}

func Authenticator(w http.ResponseWriter, r *http.Request) (*User, error) {
	cookie, err := r.Cookie("user_id")

	if err == nil {
		for _, user := range users {
			if user.Id.String() == cookie.Value {
				return &user, nil
			}
		}
	}

	http.Error(w, "", http.StatusUnauthorized)

	return nil, Unauthorized{}
}

func SetCookieAndAddUser(w http.ResponseWriter) {
	newUserId := uuid.New()

	http.SetCookie(w, &http.Cookie{
		Name:  "user_id",
		Value: newUserId.String(),
	})

	users = append(users, User{
		Id:       newUserId,
		Projects: exampleProjects,
	})
}

func main() {
	exampleTasks := []Task{
		{
			Id:          uuid.New(),
			Name:        "Complete Project Proposal",
			Description: "Draft a comprehensive project proposal outlining the scope, objectives, and deliverables. Include a timeline and resource requirements.",
			Finished:    false,
			StartDate:   nil,
		},
		{
			Id:          uuid.New(),
			Name:        "Research Market Trends",
			Description: "Conduct thorough market research to identify current trends, competitor strategies, and potential opportunities. Summarize findings in a concise report.",
			Finished:    true,
			StartDate:   nil,
		},
		{
			Id:          uuid.New(),
			Name:        "Schedule Team Meeting",
			Description: "Coordinate with team members to find a suitable time for a project status update meeting. Ensure that all key stakeholders are available and informed.",
			Finished:    false,
			StartDate:   nil,
		},
		{
			Id:          uuid.New(),
			Name:        "Review and Edit Blog Post",
			Description: "Edit a blog post draft for grammar, clarity, and style. Ensure the content aligns with the target audience and the overall content strategy.",
			Finished:    false,
			StartDate:   nil,
		},
		{
			Id:          uuid.New(),
			Name:        "Prepare Monthly Budget Report",
			Description: "Compile and analyze financial data to create a detailed budget report for the current month. Highlight any variances and provide explanations.",
			Finished:    false,
			StartDate:   nil,
		},
	}

	exampleProjects = []Project{
		{
			Id:                uuid.New(),
			Name:              "Desenvolvimento",
			UncategoriedTasks: exampleTasks,
			PrimaryColor:      "#888888",
		},
		{
			Id:                uuid.New(),
			Name:              "Cozinha",
			UncategoriedTasks: []Task{},
			PrimaryColor:      "#aa8888",
		},
	}

	projectItemTpl, _ := template.New("Project Item").Parse(`
		<button hx-target="#project-painel" hx-get="/view/project/{{.Id}}/panel" hx-swap="outerHTML" value={{.Id}} class="flex items-center w-full hover:bg-black/10 rounded px-2">
			<div style="background: {{.PrimaryColor}}" class="w-3 h-3 rounded-full mr-4"></div>
			<p class="text-2xl">{{.Name}}</p>
			<span class="text-md ml-auto">{{.UncategoriedTasks | len}}<span>
		</button>
	`)

	taskPanelTpl, _ := template.New("Task Panel").Parse(`
		<section id="project-painel" class="bg-slate-200 p-4 overflow-visible max-h-full">
			<h2 class="text-center text-3xl select-none">{{.Name}}</h2>

			<h3 class="text-2xl">Pendentes</h3>
			<ol class="list-none grid grid-cols-2 gap-4 mb-4">
				{{range $task := .UncategoriedTasks}}
					{{if not .Finished}}
						<li class="flex border-l-2 border-solid border-slate-500 p-2">
							<form>
								<div class="flex mb-2">
									<label class="checkbox_container">
										<input name="finished" type="checkbox" {{if .Finished}}checked{{end}} hx-put="/view/project/{{$.Id}}/panel/task/{{$task.Id}}" hx-target="#project-painel">
										<span class="checkmark"></span>
									</label>
									<p class="text-xl">{{$task.Name}}</p>
								</div>
								<p class="text-lg">{{$task.Description}}</p>
							</form>
						</li>
					{{end}}
				{{end}}
			</ol>

			<h3 class="text-2xl">Finalizados</h3>
			<ol class="list-none grid grid-cols-2 gap-4">
				{{range $task := .UncategoriedTasks}}
					{{if .Finished}}
						<li class="flex border-l-2 border-solid border-slate-500 p-2">
							<form>
								<div class="flex mb-2">
									<label class="checkbox_container">
										<input name="finished" type="checkbox" {{if .Finished}}checked{{end}} hx-put="/view/project/{{$.Id}}/panel/task/{{$task.Id}}" hx-target="#project-painel">
										<span class="checkmark"></span>
									</label>
									<p class="text-xl">{{$task.Name}}</p>
								</div>
								<p class="text-lg">{{$task.Description}}</p>
							</form>
						</li>
					{{end}}
				{{end}}
			</ol>
		</section>
	`)

	mux := http.NewServeMux()

	mux.HandleFunc("/view/projects", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			if user, err := Authenticator(w, r); err == nil {
				for _, project := range user.Projects {
					projectItemTpl.Execute(w, project)
				}
			}
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/view/project/", func(w http.ResponseWriter, r *http.Request) {
		if result, _ := regexp.MatchString(`^\/view\/project\/\w{8}\-\w{4}\-\w{4}\-\w{4}\-\w{12}\/panel$`, r.URL.Path); result { // no wildcard for path >:(
			switch r.Method {
			case http.MethodGet:
				if user, err := Authenticator(w, r); err == nil {
					projectIdStr := regexp.MustCompile(`\/.*`).ReplaceAllString(strings.ReplaceAll(r.URL.Path, "/view/project/", ""), "")
					for _, project := range user.Projects {
						if project.Id.String() == projectIdStr {
							taskPanelTpl.Execute(w, project)

							break
						}
					}
				}
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if result, _ := regexp.MatchString(`^\/view\/project\/.+\/panel\/task\/.+$`, r.URL.Path); result {
			switch r.Method {
			case http.MethodPut:
				formParseErr := r.ParseForm()
				if user, err := Authenticator(w, r); err == nil && formParseErr == nil {
					value := r.Form.Get("finished")
					splits := strings.Split(r.URL.Path, "/panel/task/")
					fPart := splits[0]
					sPart := splits[1]
					projectIdStr, _ := strings.CutPrefix(fPart, "/view/project/")
					taskIdStr := sPart

					for idxProject, project := range user.Projects {
						if project.Id.String() == projectIdStr {
							for idxTask, task := range project.UncategoriedTasks {
								if task.Id.String() == taskIdStr {
									if value == "on" {
										user.Projects[idxProject].UncategoriedTasks[idxTask].SetFinished(true)
									} else if value == "" {
										user.Projects[idxProject].UncategoriedTasks[idxTask].SetFinished(false)
									}

									break
								}
							}

							taskPanelTpl.Execute(w, project)

							break
						}
					}
				}
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("user_id")

		if err == http.ErrNoCookie {
			SetCookieAndAddUser(w)
		} else {
			userExists := false
			for _, user := range users {
				if user.Id.String() == cookie.Value {
					userExists = true

					break
				}
			}

			if !userExists {
				SetCookieAndAddUser(w)
			}
		}

		http.ServeFile(w, r, "./page/index.html")
	})

	mux.HandleFunc("/static/", func(w http.ResponseWriter, r *http.Request) {
		filename, _ := strings.CutPrefix(r.URL.Path, "/static/")
		if file, err := http.Dir("./static").Open(filename); err == nil {
			if strings.HasSuffix(filename, ".js") {
				w.Header().Set("Content-Type", "text/javascript")
			} else if strings.HasSuffix(filename, ".css") {
				w.Header().Set("Content-Type", "text/css")
			}

			bytes := make([]byte, 4)
			for {
				size, err := file.Read(bytes)
				if err == nil {
					w.Write(bytes[:size])
				} else if err == io.EOF {
					break
				} else {
					break
				}
			}
		} else {
			http.Error(w, "Not found", http.StatusNotFound)
		}
	})

	log.Fatal(http.ListenAndServe(":3050", mux))
}
