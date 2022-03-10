package dag

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"sync"

	"github.com/golang/protobuf/jsonpb"
	pdeployment "github.com/n0stack/n0stack/n0proto.go/deployment/v0"
	ppool "github.com/n0stack/n0stack/n0proto.go/pool/v0"
	pprovisioning "github.com/n0stack/n0stack/n0proto.go/provisioning/v0"

	"github.com/fatih/color"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
)

var Marshaler = &jsonpb.Marshaler{
	EnumsAsInts:  true,
	EmitDefaults: false,
	OrigName:     true,
}

type Task struct {
	Type        string      `yaml:"type"`
	Action      string      `yaml:"action"`
	Args        interface{} `yaml:"args"`
	DependsOn   []string    `yaml:"depends_on"`
	IgnoreError bool        `yaml:"ignore_error"`
	// Rollback []*Task `yaml:"rollback"`

	child   []string
	depends int
}

// referenced by https://stackoverflow.com/questions/40737122/convert-yaml-to-json-without-struct
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m2 := map[string]interface{}{}
		for k, v := range x {
			m2[k.(string)] = convert(v)
		}
		return m2
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}

	return i
}

// TODO :生成自動化
func GetMessageType(s string, conn *grpc.ClientConn) (reflect.Type, reflect.Value, bool) {
	var t reflect.Type
	var v reflect.Value

	switch s {
	case "node", "Node":
		t = reflect.TypeOf(ppool.NewNodeServiceClient(conn))
		v = reflect.ValueOf(ppool.NewNodeServiceClient(conn))
	case "network", "Network":
		t = reflect.TypeOf(ppool.NewNetworkServiceClient(conn))
		v = reflect.ValueOf(ppool.NewNetworkServiceClient(conn))
	case "block_storage", "BlockStorage":
		t = reflect.TypeOf(pprovisioning.NewBlockStorageServiceClient(conn))
		v = reflect.ValueOf(pprovisioning.NewBlockStorageServiceClient(conn))
	case "virtual_machine", "VirtualMachine":
		t = reflect.TypeOf(pprovisioning.NewVirtualMachineServiceClient(conn))
		v = reflect.ValueOf(pprovisioning.NewVirtualMachineServiceClient(conn))
	case "image", "Image":
		t = reflect.TypeOf(pdeployment.NewImageServiceClient(conn))
		v = reflect.ValueOf(pdeployment.NewImageServiceClient(conn))
	default:
		return nil, v, false
	}

	return t, v, true
}

// return response JSON bytes
func (a Task) Do(conn *grpc.ClientConn) (proto.Message, error) {
	grpcCliType, grpcCliValue, ok := GetMessageType(a.Type, conn)
	if !ok {
		return nil, fmt.Errorf("Resource type '%s' do not exist", a.Type)
	}

	fnt, ok := grpcCliType.MethodByName(a.Action)
	if !ok {
		return nil, fmt.Errorf("Resource type '%s' do not have action '%s'", a.Type, a.Action)
	}

	// 1st arg is instance, 2nd is context.Background()
	// TODO: 何かがおかしい、 argsElem is "**SomeMessage", so use argsElem.Elem() in Call
	argsType := fnt.Type.In(2)
	argsElem := reflect.New(argsType)
	if a.Args == nil {
		a.Args = make(map[string]interface{})
	}
	buf, err := json.Marshal(convert(a.Args))
	if err != nil {
		return nil, fmt.Errorf("Args is invalid, set fields of message '%s' err=%s", argsType.String(), err.Error())
	}
	if err := json.Unmarshal(buf, argsElem.Interface()); err != nil {
		return nil, fmt.Errorf("Args is invalid, set fields of message '%s' err=%s", argsType.String(), err.Error())
	}

	out := fnt.Func.Call([]reflect.Value{grpcCliValue, reflect.ValueOf(context.Background()), argsElem.Elem()})
	if err, _ := out[1].Interface().(error); err != nil {
		return nil, fmt.Errorf("got error response: %s", err.Error())
	}

	return out[0].Interface().(proto.Message), nil
}

// topological sort
// 実際遅いけどもういいや O(E^2 + V)
func CheckDAG(tasks map[string]*Task) error {
	result := 0

	for k := range tasks {
		tasks[k].child = make([]string, 0)
		tasks[k].depends = len(tasks[k].DependsOn)
	}

	for k, v := range tasks {
		for _, d := range v.DependsOn {
			if _, ok := tasks[d]; !ok {
				return fmt.Errorf("Depended task '%s' do not exist", d)
			}

			tasks[d].child = append(tasks[d].child, k)
		}
	}

	s := make([]string, 0, len(tasks))
	for k, v := range tasks {
		if v.depends == 0 {
			s = append(s, k)
			result++
		}
	}

	for len(s) != 0 {
		n := s[len(s)-1]
		s = s[:len(s)-1]

		for _, c := range tasks[n].child {
			tasks[c].depends--
			if tasks[c].depends == 0 {
				s = append(s, c)
				result++
			}
		}
	}

	if result != len(tasks) {
		return fmt.Errorf("This request is not DAG")
	}

	return nil
}

type ActionResult struct {
	Name string
	Res  proto.Message
	Err  error
}

// 出力で時間を出したほうがよさそう
func DoDAG(ctx context.Context, tasks map[string]*Task, out io.Writer, conn *grpc.ClientConn) bool {
	for k := range tasks {
		tasks[k].child = make([]string, 0)
		tasks[k].depends = len(tasks[k].DependsOn)
	}

	for k, v := range tasks {
		for _, d := range v.DependsOn {
			tasks[d].child = append(tasks[d].child, k)
		}
	}

	resultChan := make(chan ActionResult, 100)
	wg := new(sync.WaitGroup)
	total := len(tasks)
	done := 0

	doTask := func(taskName string) {
		defer wg.Done()

		result, err := tasks[taskName].Do(conn)
		resultChan <- ActionResult{
			Name: taskName,
			Res:  result,
			Err:  err,
		}
	}

	for k, v := range tasks {
		if v.depends == 0 {
			wg.Add(1)
			fmt.Fprintf(out, "---> Task '%s' is started\n", k)
			log.Printf("[DEBUG] Task '%s' is started: %+v", k, v)
			go doTask(k)
		}
	}

	canceled := false
	failed := false

	go func() {
		<-ctx.Done()

		canceled = true

		fmt.Println("---> Wait to finish requested tasks")
		if !failed {
			go func() {
				wg.Wait()
				close(resultChan)
			}()
		}
	}()

	for r := range resultChan {
		done++

		if r.Err != nil {
			if tasks[r.Name].IgnoreError {
				color.Set(color.FgRed)
				fmt.Fprintf(out, "---> [ %d/%d ] Task '%s' is failed, ignore error: ", done, total, r.Name)
				color.Unset()
				color.Set(color.FgWhite)
				fmt.Fprintf(out, "%s\n", r.Err.Error())
				color.Unset()
			} else {
				color.Set(color.FgRed)
				fmt.Fprintf(out, "---> [ %d/%d ] Task '%s' is failed: ", done, total, r.Name)
				color.Unset()
				color.Set(color.FgWhite)
				fmt.Fprintf(out, "%s\n", r.Err.Error())
				color.Unset()

				if !failed {
					failed = true

					// すでにリクエストしたタスクの終了を待つ
					fmt.Fprintf(out, "---> Wait to finish requested tasks\n")
					if !canceled {
						go func() {
							wg.Wait()
							close(resultChan)
						}()
					}
				}
			}
		} else {
			res, _ := Marshaler.MarshalToString(r.Res)

			if failed {
				color.Set(color.Attribute(91))
				fmt.Fprintf(out, "---> [ %d/%d ] Task '%s', which was requested until failed, is finished\n", done, total, r.Name)
				color.Unset()
				color.Set(color.FgWhite)
				fmt.Fprintf(out, "%s\n\n", res)
				color.Unset()
			} else {
				color.Set(color.FgGreen)
				fmt.Fprintf(out, "---> [ %d/%d ] Task '%s' is finished\n", done, total, r.Name)
				color.Unset()
				color.Set(color.FgWhite)
				fmt.Fprintf(out, "%s\n\n", res)
				color.Unset()
			}
		}

		if !(failed || canceled) {
			// queueing
			for _, d := range tasks[r.Name].child {
				tasks[d].depends--
				if tasks[d].depends == 0 {
					wg.Add(1)
					fmt.Fprintf(out, "---> Task '%s' is started\n", d)
					log.Printf("[DEBUG] Task '%s' is started: %+v", d, tasks[d])
					go doTask(d)
				}
			}
		}

		if !(failed || canceled) && done == total {
			close(resultChan)
		}
	}

	if failed || (canceled && done != total) {
		// TODO: rollback
		return false
	}

	return true
}
