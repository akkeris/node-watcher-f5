package utils



import (
     "fmt"
     "github.com/jasonlvhit/gocron"
     "os"
     "strconv"
 )



func StartTokenCron() {
        s := gocron.NewScheduler()
        t := getEnvUint64("AUTH_TOKEN_RETRIEVE_TIME_IN_MINS")
        s.Every(t).Minutes().Do(NewToken)
        <-s.Start()
}


func getEnvInt(s string) int {
        v := os.Getenv(s)
        i, err := strconv.Atoi(v)
        if err != nil {
                fmt.Println(s, " must be an integer.")
        }
        return i
}


func getEnvUint64(s string) uint64 {
        i := getEnvInt(s)
        if i < 0 {
                fmt.Println(s, " must be a positive integer.")
        }
        return uint64(i)
}

