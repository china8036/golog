## Intruction
* 日志先写入内存中后定时刷入文件 以减少IO消耗


## Install


```sh
go get  github.com/china8036/log
```

## Example

```go
package main

import (
	"time"

	"github.com/china8036/golang/log"
)

func main() {
	for {
		time.Sleep(time.Second)
		log.LogError(time.Now().String())
	}

	time.Sleep(time.Minute)
}


```

