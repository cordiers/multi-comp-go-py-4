package dag

import (
	"testing"

	"google.golang.org/grpc"
)

// func TestHoge(t *testing.T) {
// 	a := struct {
// 		Hoge string
// 		Foo  string
// 	}{
// 		"hage",
// 		"bar",
// 	}

// 	va := reflect.ValueOf(a)
// 	buf, err := json.Marshal(va.Interface())
// 	if err != nil {
// 		t.Errorf("err: %+v", err.Error())
// 	}

// 	t.Errorf("buf: %s", string(buf))
// }

func TestIsDAG(t *testing.T) {
	cases := []struct {
		name    string
		actions map[string]*Task
		err     string
	}{
		{
			"loop",
			map[string]*Task{
				"g1": {
					DependsOn: []string{
						"g2",
					},
				},
				"g2": {
					DependsOn: []string{
						"g3",
					},
				},
				"g3": {
					DependsOn: []string{
						"g1",
					},
				},
			},
			"This request is not DAG",
		},
		{
			"liner",
			map[string]*Task{
				"g1": {
					DependsOn: []string{
						"g2",
					},
				},
				"g2": {
					DependsOn: []string{
						"g3",
					},
				},
				"g3": {},
			},
			"",
		},
	}

	for _, c := range cases {
		err := CheckDAG(c.actions)

		if (c.err == "") == (err != nil) {
			t.Errorf("[%s] wrong existence error: have=%+v, want=%+v", c.name, err != nil, c.err == "")
		}
		if c.err != "" && err.Error() != c.err {
			t.Errorf("[%s] got wrong err: have=%s, want='%s'", c.name, err.Error(), c.err)
		}
	}
}

func TestDoDAG(t *testing.T) {
	endpoint := "localhost:20180"
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to connect API: endpoint=%s, err=%s", endpoint, err.Error())
	}
	defer conn.Close()

	cases := []struct {
		name string
		task *Task
	}{
		{
			"loop",
			&Task{
				Type:   "Node",
				Action: "ApplyNode",
				Args: map[string]interface{}{
					"name": "mock-node",
					"annotations": map[interface{}]interface{}{
						"test": "test",
					},
				},
				DependsOn: []string{
					"g3",
				},
			},
		},
	}

	for _, c := range cases {
		c.task.Do(conn)
	}
}

// func TestDoDAG(t *testing.T) {
// 	endpoint := "localhost:20180"
// 	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
// 	if err != nil {
// 		t.Fatalf("Failed to connect API: endpoint=%s, err=%s", endpoint, err.Error())
// 	}
// 	defer conn.Close()

// 	cases := []struct {
// 		name    string
// 		actions map[string]*Task
// 		result  bool
// 	}{
// 		{
// 			"loop",
// 			map[string]*Task{
// 				"g1": &Task{
// 					DependOn: []string{
// 						"g2",
// 					},
// 				},
// 				"g2": &Task{
// 					DependOn: []string{
// 						"g3",
// 					},
// 				},
// 				"g3": &Task{
// 					DependOn: []string{
// 						"g1",
// 					},
// 				},
// 			},
// 			false,
// 		},
// 		{
// 			"liner",
// 			map[string]*Task{
// 				"g1": &Task{
// 					DependOn: []string{
// 						"g2",
// 					},
// 				},
// 				"g2": &Task{
// 					DependOn: []string{
// 						"g3",
// 					},
// 				},
// 				"g3": &Task{},
// 			},
// 			true,
// 		},
// 	}

// 	for _, c := range cases {
// 		if DoDAG(c.actions) != c.result {
// 			t.Errorf("[%s]", c.name)
// 		}
// 	}
// }
