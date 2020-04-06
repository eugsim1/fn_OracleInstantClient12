
package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"time"
	"os"
    "fmt"
	"os/exec"
	"path/filepath"
	fdk "github.com/fnproject/fdk-go"
	"github.com/oracle/oci-go-sdk/objectstorage"
	"github.com/oracle/oci-go-sdk/common/auth"
	"github.com/oracle/oci-go-sdk/common"
)

var OpFileName string

func main() {
	fdk.Handle(fdk.HandlerFunc(atpEventCreation))
	OpFileName = "debug.txt"
}

func atpEventCreation(ctx context.Context, in io.Reader, out io.Writer) {
	log.Println("ATP/AWD DB events notification handler invoked on", time.Now())
    OpFileName = "debug.txt"
	var evt OCIEvent

	f, err := os.Create("/tmp/"+OpFileName)
	if err != nil {
        log.Printf("cant create local log file\n",err)
		resp := ResptoCallerError{Message: "cant create local log file", Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
        return
    }	
	
	json.NewDecoder(in).Decode(&evt)
	log.Println("Got OCI cloud event payload")
	
	data, _ := json.Marshal(evt)
    jsonStr := string(data)
	fmt.Fprintf(out,"%v\n", jsonStr)
	_, err = f.WriteString(jsonStr )
	

	eventDetails := evt.Data
	log.Println("Event data", eventDetails)
	data, _ = json.Marshal(eventDetails)
    jsonStr = string(data)
	fmt.Fprintf(out,"json string %v\n", jsonStr)
	_, err = f.WriteString(jsonStr )
	
	Add_Detail := eventDetails.AdditionalDetails
	log.Println("Event Add_Detail", Add_Detail)
	data, _ = json.Marshal(Add_Detail)
    jsonStr = string(data)
	fmt.Fprintf(out,"json string %v\n", jsonStr)
	_, err = f.WriteString(jsonStr )
	
 	 
    test_adw(ctx,out,Add_Detail.DbName, f )
	response := "ok"
	log.Println("Response", "ok")
	fmt.Fprintf(out,"Response ok")
	out.Write([]byte(response))
}

// cant call sqlplus with the go 
// as i m getting TNS protocol adapter errors

func test_adw(ctx context.Context, out io.Writer,Dbname string, f *os.File) {

    os.Setenv("TNS_ADMIN", "/function/wallet")
	os.Setenv("PATH","/usr/lib/oracle/12.2/client64/lib:/usr/sbin:/usr/bin:/function:/go/bin:/usr/local/go/bin")
	os.Setenv("LD_LIBRARY_PATH", "/usr/lib/oracle/12.2/client64/lib")
	os.Setenv("SHELL","/bin/bash")
	
  path, errpath :=exec.LookPath("sqlplus")
  if errpath != nil {
		fmt.Fprintf(out,"sqlplus not in path")
	}
	fmt.Fprintf(out,"\nsqlplus is available at %s\n", path)
	

 out1, err := exec.Command("/bin/bash", "-c", "/function/test_sql.sh").Output()
    if err != nil {
        fmt.Fprintf(out, "sqlplus error =>%s\n", err)
    } else {
	fmt.Println("Command Successfully Executed")
    output := string(out1[:])
    fmt.Fprintf(out,"sqlplus \n%s\n",output)
	}


 UploadtoObjectStorage(ctx , out , "adwfree_after", "/tmp/cli.trc", f)	
}	 


func UploadtoObjectStorage(ctx context.Context, out io.Writer, Prefix string, OpFileName string,  f *os.File ) {
//time_now := time.Now().Add(time.Hour * -1).Format(time.RFC3339)

    fnctx := fdk.GetContext(ctx)
	outputBucket := fnctx.Config()["OUTPUT_BUCKET"]
	namespace := fnctx.Config()["NAMESPACE"]
	
	rp, err := auth.ResourcePrincipalConfigurationProvider()
	if err != nil {
	    fmt.Fprintf(out, "{\"Audif_info\": %s}\n", err)
		fmt.Fprintf(f, "{\"Audif_info\": %s}\n", err)
	    log.Printf("Error %v", err)
		panic(err)
	}

	ObjectStorageClient, err := objectstorage.NewObjectStorageClientWithConfigurationProvider(rp)
		if err != nil {
		resp := ResptoCallerError{Message: "Problem getting Object Store Client handle", Error: err.Error()}
		log.Println(resp.toString())
		
		_, err = f.WriteString(resp.toString() )
		f.Close()
		
		
		json.NewEncoder(out).Encode(resp)
		return
	}

	info, err := os.Stat(OpFileName)
	
		if err != nil {
		msg := "error on f.Stat() call"
		resp := ResptoCallerError{Message: "error on f.Stat() call OpFileName", Error: err.Error()}
		log.Println(msg)
		json.NewEncoder(out).Encode(resp)
		fmt.Fprintf(f, "{\"error on f.Stat() call OpFileName\": %s %s}\n", err, OpFileName)
		return
	} 
	
	file, err := os.Open(OpFileName)
		if err != nil {
		msg := "error on os.Open() call"
		resp := ResptoCallerError{Message: "error on os.Open call loc_debug_filename", Error: err.Error()}
		log.Println(msg)
		json.NewEncoder(out).Encode(resp)
		fmt.Fprintf(f, "{\"error on f.Stat() call OpFileName\": %s %s}\n", err, OpFileName)
		return
	} 
	
	defer file.Close()
	
	path := OpFileName
    fileUpload := filepath.Base(path)
    fmt.Fprintf(f, "{\"File to upload\": %s}\n",Prefix+fileUpload)

	putReq := objectstorage.PutObjectRequest{ NamespaceName: common.String(namespace), BucketName: common.String(outputBucket), ObjectName: common.String(Prefix+fileUpload),ContentLength: common.Int64(info.Size()), PutObjectBody: file }
	_, err = ObjectStorageClient.PutObject(context.Background(), putReq)

	if err == nil {
		msg := "logfile " + OpFileName + " written to storage bucket - " + outputBucket
		log.Println(msg)
		//out.Write([]byte(msg))
	} else {
		resp := ResptoCallerError{Message: "Failed to write to bucket", Error: err.Error()}
		log.Println(resp.toString())
		json.NewEncoder(out).Encode(resp)
		return
	}
}

//OCIEvent structure
type OCIEvent struct {
	EventType          string `json:"eventType"`
	CloudEventsVersion string `json:"cloudEventsVersion"`
	Source             string `json:"source"`
	EventID            string `json:"eventID"`
	Data               Data   `json:"data"`
	EventTypeVersion   string `json:"eventTypeVersion"`
	EventTime          string `json:"eventTime"`
	SchemaURL          string `json:"schemaURL"`
	ContentType        string `json:"contentType"`
	Extensions         Extensions `json:"extensions"`

}

//Extensions - "extension" attribute in events JSON payload
type Extensions struct {
	CompartmentId string `json:"compartmentId"`
}

//Data - represents (part of) the event data
type Data struct {
	ID             string `json:"compartmentId"`
	ResourceName   string `json:"resourceName"`
	ResourceId     string `json:"resourceId"`
	AdditionalDetails AdditionalDetails  `json:"additionalDetails"`
}

type AdditionalDetails struct {
CpuCoreCount int `json:"cpuCoreCount"`
LifecycleState string `json:"lifecycleState"`
DbName string `json:"dbName"`
AutonomousDatabaseId string `json:"autonomousDatabaseId"`
}


type ResptoCallerError struct {
	Message string
	Error   string
}


func (response ResptoCallerError) toString() string {
	return response.Message + " due to " + response.Error
}



 
 
