package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"strconv"
)

//
// Map functions return a slice of KeyValue.
//
type KeyValue struct {
	Key   string
	Value string
}
type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

//
// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
//
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

//
// main/mrworker.go calls this function.
//
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string) {

	// Your worker implementation here.

	// uncomment to send the Example RPC to the master.

	var filename string
	var workernum int
	var worktype string

	var intermediatefile string

	for {
		filename, workernum, worktype = CallRequestTask()
		if filename == "" {
			os.Exit(1)
		}
		if worktype == "Mapper" {
			fmt.Println("Got a map task from master with filename = ", filename)
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}
			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", filename)
			}
			file.Close()

			kva := mapf(filename, string(content))

			// The following logic creates files of type mapper-X-Y
			// X = workernum
			// Y = Final Intermediate file
			fileMap := make(map[string]*os.File)
			for i := 0; i < 10; i++ {
				fName := "mapper" + "-" + strconv.Itoa(workernum) + "-" + strconv.Itoa(i)
				if _, err := os.Stat(fName); err == nil {
					fileMap[fName], err = os.OpenFile(fName, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
					if err != nil {
						fmt.Println(err)
					}
				} else if os.IsNotExist(err) {
					fileMap[fName], err = os.Create(fName)
					if err != nil {
						fmt.Println(err)
					}
				} else {
					fmt.Println("Something else is going on !")
				}
			}

			var FileSet = make(map[string]bool)
			for _, kv := range kva {
				filenumber := ihash(kv.Key) % 10
				intermediatefile = "mapper" + "-" + strconv.Itoa(workernum) + "-" + strconv.Itoa(filenumber)
				if _, ok := FileSet[intermediatefile]; !ok {
					// intermediatefile does not exists in the SET of Intermediate Files
					// Therefore Insert this intermediate file in the SET
					FileSet[intermediatefile] = true
				}
				enc := json.NewEncoder(fileMap[intermediatefile])
				err = enc.Encode(&kv)
			}
			for _, fp := range fileMap {
				fp.Close()
			}
			//notify master that Mapper has finished its task
			intermediatefileList := make([]string, 0)
			for k, _ := range FileSet {
				intermediatefileList = append(intermediatefileList, k)
			}
			fmt.Println("Number of files sent to master = ", len(intermediatefileList)) // output should be <= 10
			mRequest := MapperRequest{intermediatefileList, 2, filename}                // 2 means mapper task is done
			CallMapperDone(mRequest)
			fmt.Println("Mapper Task Done! :", workernum)

		} else {
			// call reducef
			fmt.Println("Got a reduce task from master with filename = ", filename)
			intermediatefile = filename
			intermediatekva := []KeyValue{}
			x, err := os.Open(intermediatefile)
			if err != nil {
				fmt.Println(err)
			}
			dec := json.NewDecoder(x)
			for {
				var kv KeyValue
				if err = dec.Decode(&kv); err != nil {
					break
				}
				intermediatekva = append(intermediatekva, kv)
			}
			x.Close()

			sort.Sort(ByKey(intermediatekva))

			oname := "mr-out-" + strconv.Itoa(workernum)
			ofile, _ := os.Create(oname)

			//
			// call Reduce on each distinct key in intermediate[],
			// and print the result to mr-out-0.
			//
			i := 0
			for i < len(intermediatekva) {
				j := i + 1
				for j < len(intermediatekva) && intermediatekva[j].Key == intermediatekva[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediatekva[k].Value)
				}
				output := reducef(intermediatekva[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(ofile, "%v %v\n", intermediatekva[i].Key, output)

				i = j
			}

			ofile.Close()
			// notify master that reducer has finished its job
			rReq := ReducerRequest{oname, 1}
			CallReducerDone(rReq)
		}
	}

}

//
// example function to show how to make an RPC call to the master.
//
// the RPC argument and reply types are defined in rpc.go.
//
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	call("Master.Example", &args, &reply)

	// reply.Y should be 100.
	fmt.Printf("reply.Y %v\n", reply.Y)
}

func CallRequestTask() (string, int, string) {

	// declare an argument structure.
	req := MrRequest{}

	// declare a reply structure.
	reply := MrReply{}

	// send the RPC request, wait for the reply.
	call("Master.RequestTask", &req, &reply)

	return reply.FileName, reply.WorkerNum, reply.WorkType
}

func CallMapperDone(req MapperRequest) {

	// declare an argument structure.
	// req := MrRequest{}

	// declare a reply structure.
	reply := MrEmpty{}

	// send the RPC request, wait for the reply.
	call("Master.MapperDone", &req, &reply)
}

func CallReducerDone(req ReducerRequest) {

	// declare an argument structure.
	// req := MrRequest{}

	// declare a reply structure.
	reply := MrEmpty{}

	// send the RPC request, wait for the reply.
	call("Master.ReducerDone", &req, &reply)
}

//
// send an RPC request to the master, wait for the response.
// usually returns true.
// returns false if something goes wrong.
//
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	c, err := rpc.DialHTTP("unix", "mr-socket")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	fmt.Println(err)
	return false
}
