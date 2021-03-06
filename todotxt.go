package todotxt

import (
        "time"
        "os"
        "bufio"
        "strings"
        "regexp"
        "sort"
        "unicode"
        "fmt"
        "math/rand"
)

type Task struct {
        id int
        todo string
        priority byte
        create_date time.Time
        contexts []string
        projects []string
        raw_todo string
        finished bool
        finish_date time.Time
        id_padding int
}

type TaskList []Task

func ParseTask(text string, id int) (Task) {
        var task = Task{}
        task.id = id
        task.raw_todo = text

        splits := strings.Split(text, " ")

        // checking if the task is already finished
        if text[0] == 'x' &&
           text[1] == ' ' &&
           !unicode.IsSpace(rune(text[2])) {
                task.finished = true
                splits = splits[1:]
        }

        date_regexp := "([\\d]{4})-([\\d]{2})-([\\d]{2})"

        // checking for finish date
        if match, _ := regexp.MatchString(date_regexp, splits[0]); match {
                if date, e := time.Parse("2006-01-02", splits[0]); e != nil {
                        panic(e)
                } else {
                        task.finish_date = date
                }

                splits = splits[1:]
        }

        head := splits[0]

        // checking for priority
        if (len(head) == 3) &&
           (head[0] == '(') &&
           (head[2] == ')') &&
           (head[1] >= 65 && head[1] <= 90) { // checking if it's in range [A-Z]
                task.priority = head[1]
                splits = splits[1:]
        }

        // checking for creation date and building the actual todo item
        if match, _ := regexp.MatchString(date_regexp, splits[0]); match {
                if date, e := time.Parse("2006-01-02", splits[0]); e != nil {
                        panic(e)
                } else {
                        task.create_date = date
                }

                task.todo = strings.Join(splits[1:], " ")
        } else {
                task.todo = strings.Join(splits[0:], " ")
        }

        context_regexp, _ := regexp.Compile("@[[:word:]]+")
        contexts := context_regexp.FindAllStringSubmatch(text, -1)
        if len(contexts) != 0 {
                for _, context := range contexts {
                        task.contexts = append(task.contexts, context[0])
                }
        }

        project_regexp, _ := regexp.Compile("\\+[[:word:]]+")
        projects := project_regexp.FindAllStringSubmatch(text, -1)
        if len(projects) != 0 {
                for _, project := range projects {
                        task.projects = append(task.projects, project[0])
                }
        }

        return task
}

func LoadTaskList (filename string) (TaskList) {

        var f, err = os.Open(filename)

        if err != nil {
                panic(err)
        }

        defer f.Close()

        var tasklist = TaskList{}

        scanner := bufio.NewScanner(f)

        for scanner.Scan() {
                text := scanner.Text()
                tasklist.Add(text)
        }

        if err := scanner.Err(); err != nil {
                panic(scanner.Err())
        }

        return tasklist
}

type By func(t1, t2 Task) bool

func (by By) Sort(tasks TaskList) {
        ts := &taskSorter{
                tasks: tasks,
                by:    by,
        }
        sort.Sort(ts)
}

type taskSorter struct {
        tasks TaskList
        by func(t1, t2 Task) bool
}

func (s *taskSorter) Len() int {
        return len(s.tasks)
}

func (s *taskSorter) Swap(i, j int) {
        s.tasks[i], s.tasks[j] = s.tasks[j], s.tasks[i]
}

func (s *taskSorter) Less(i, j int) bool {
        return s.by(s.tasks[i], s.tasks[j])
}

func (tasks TaskList) Len() int {
        return len(tasks)
}

func prioCmp(t1, t2 Task) bool {
        return t1.Priority() < t2.Priority()
}

func prioRevCmp(t1, t2 Task) bool {
        return t1.Priority() > t2.Priority()
}

func dateCmp(t1, t2 Task) bool {
        tm1 := t1.CreateDate().Unix()
        tm2 := t2.CreateDate().Unix()

        // if the dates equal, let's use priority
        if tm1 == tm2 {
                return prioCmp(t1, t2)
        } else {
                return tm1 > tm2
        }
}

func dateRevCmp(t1, t2 Task) bool {
        tm1 := t1.CreateDate().Unix()
        tm2 := t2.CreateDate().Unix()

        // if the dates equal, let's use priority
        if tm1 == tm2 {
                return prioCmp(t1, t2)
        } else {
                return tm1 < tm2
        }
}

func lenCmp(t1, t2 Task) bool {
        tl1 := len(t1.raw_todo)
        tl2 := len(t2.raw_todo)
        if tl1 == tl2 {
                return prioCmp(t1, t2)
        } else {
                return tl1 < tl2
        }
}

func lenRevCmp(t1, t2 Task) bool {
        tl1 := len(t1.raw_todo)
        tl2 := len(t2.raw_todo)
        if tl1 == tl2 {
                return prioCmp(t1, t2)
        } else {
                return tl1 > tl2
        }
}

func idCmp(t1, t2 Task) bool {
        return t1.Id() < t2.Id()
}

func randCmp(t1, t2 Task) bool {
        rand.Seed(time.Now().UnixNano()%1e6/1e3)
        return rand.Intn(len(t1.raw_todo)) > rand.Intn(len(t2.raw_todo))
}

func (tasks TaskList) Sort(by string) {
        switch by {
        default:
        case "prio":
                By(prioCmp).Sort(tasks)
        case "prio-rev":
                By(prioRevCmp).Sort(tasks)
        case "date":
                By(dateCmp).Sort(tasks)
        case "date-rev":
                By(dateRevCmp).Sort(tasks)
        case "len":
                By(lenCmp).Sort(tasks)
        case "len-rev":
                By(lenRevCmp).Sort(tasks)
        case "id":
                By(idCmp).Sort(tasks)
        case "rand":
                By(randCmp).Sort(tasks)
        }
}

func (tasks TaskList) Save(filename string) {
        tasks.Sort("id")

        f, err := os.Create(filename)
        if err != nil {
                panic(err)
        }

        defer f.Close()

        for _, task := range tasks {
                f.WriteString(task.RawText() + "\n")
        }
        f.Sync()
}

func (tasks *TaskList) Add(todo string) {
        task := ParseTask(todo, tasks.Len())
        *tasks = append(*tasks, task)
}

func (tasks TaskList) Done(id int, finish_date bool) error {
        if id > tasks.Len() || id < 0 {
                return fmt.Errorf("Error: id is %v", id)
        }

        tasks[id].finished = true
        if finish_date {
                t := time.Now()
                tasks[id].raw_todo = "x " + t.Format("2006-01-02") + " " +
                                        tasks[id].raw_todo
        } else {
                tasks[id].raw_todo = "x " + tasks[id].raw_todo
        }

        return nil
}

func (task Task) Id() int {
        return task.id
}

func (task Task) Text() string {
        return task.todo
}

func (task Task) RawText() string {
        return task.raw_todo
}

func (task Task) Priority() byte {
        // if priority is not from [A-Z], let it be 94 (^)
        if task.priority < 65 || task.priority > 90 {
                return 94 // you know, ^
        } else {
                return task.priority
        }
}

func (task Task) Contexts() []string {
        return task.contexts
}

func (task Task) Projects() []string {
        return task.projects
}

func (task Task) CreateDate() time.Time {
        return task.create_date
}

func (task Task) Finished() bool {
        return task.finished
}

func (task Task) FinishDate() time.Time {
        return task.finish_date
}

func (task *Task) SetIdPaddingBy(tasklist TaskList) {
        l := tasklist.Len()

        if l >= 10000 {
                task.id_padding = 5
        } else if l >= 1000 {
                task.id_padding = 4
        } else if l >= 100 {
                task.id_padding = 3
        } else if l >= 10 {
                task.id_padding = 2
        } else {
                task.id_padding = 1
        }
}

func (task *Task) RebuildRawTodo() {
        if task.finished {
                task.raw_todo = task.PrettyPrint("x %P%t")
        } else {
                task.raw_todo = task.PrettyPrint("%P%t")
        }
}

func (task *Task) SetPriority(prio byte) {
        if prio < 65 || prio > 90 {
                task.priority = '^'
        } else {
                task.priority = prio
        }
}

func (task *Task) SetTodo(todo string) {
        task.todo = todo
}

func (task Task) IdPadding() int {
        return task.id_padding
}

func (task Task) ANSIColor() string {
        if task.Priority() == 65 { // A
                return "\033[1;31m" //red
        } else if task.Priority() == 66 { // B
                return "\033[1;33m" //yellow
        } else if task.Priority() == 67 { // C
                return "\033[0;32m" //green
        } else if task.Priority() == 68 { // C
                return "\033[1;34m" //blue
        } else {
                return "\033[0m"
        }
}

func pad(in string, length int) string {
        if (length == -1) {
                return in
        }

        if (length > len(in)) {
                return strings.Repeat(" ", length - len(in)) + in
        } else {
                return in[:length]
        }
}

func soft_pad(in string, length int) string {
        if (length == -1) {
                return in
        }

        if (length > len(in)) {
                return strings.Repeat(" ", length - len(in)) + in
        } else {
                slice := strings.Split(in, " ")
                return soft_pad(strings.Join(slice[:len(slice)-1], " "), length)
        }
}

func (task Task) PrettyPrint(pretty string) string {
        rp := regexp.MustCompile("(%((\\.|\\*)\\d+|)[a-zA-Z])")
        padding := -1
        out := rp.ReplaceAllStringFunc(pretty, func(s string) string {
                if (len(s) < 2) {
                        return ""
                }

                var f string
                if (s[0] == '%' && (s[1] == '.' || s[1] == '*')) {
                        if _, e := fmt.Sscanf(s[2:], "%d%s", &padding, &f); e != nil {
                                panic(e);
                        }
                        f = "%" + f
                } else {
                        f = s
                }

                ret := s
                switch f {
                case "%i":
                        str := fmt.Sprintf("%%0%dd", task.IdPadding())
                        ret = fmt.Sprintf(str, task.Id())
                case "%t":
                        ret = task.Text()
                case "%T":
                        ret = task.RawText()
                case "%p":
                        ret = string(task.Priority())
                case "%P":
                        if task.Priority() != '^' {
                                ret = "(" + string(task.Priority()) + ") "
                        } else {
                                ret = ""
                        }
                case "%d":
                        ret = fmt.Sprintf("%d", task.CreateDate().Day())
                case "%m":
                        ret = fmt.Sprintf("%d", int(task.CreateDate().Month()))
                case "%y":
                        ret = fmt.Sprintf("%d", task.CreateDate().Year())

                case "%D":
                        ret = fmt.Sprintf("%d", task.FinishDate().Day())
                case "%M":
                        ret = fmt.Sprintf("%d", int(task.FinishDate().Month()))
                case "%Y":
                        ret = fmt.Sprintf("%d", task.FinishDate().Year())
                case "%c":
                        ret = fmt.Sprintf("%s", task.ANSIColor())
                case "%r":
                        ret = fmt.Sprintf("\033[0m") // ANSI color reset


                default:
                        ret = s
                }

                switch s[1] {
                case '.':
                        ret = soft_pad(ret, padding)
                case '*':
                        ret = pad(ret, padding)
                }

                padding = -1
                return ret
        })
        return out
}

func (task Task) Matches(text string) bool {
        match, _ := regexp.MatchString(text, task.Text())
        return match
}
