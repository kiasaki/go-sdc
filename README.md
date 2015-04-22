# go-sdc

`go-sdc` is a Go client library for accessing the SDC 7.0 API.

_SDC stands for Smart Data Center and is a product by **Joyent** powering their public and private cloud offering._

## Usage

```go
import "github.com/kiasaki/go-sdc"
```

Create a new SDC client, then use it's exposed methods to query the SDC API.

## Endpoints implemented

### GetMachine `GET (/:login/machines/:id)`

### CreateMachine `POST (/:login/machines)`

## Credits

Goods parts of the Client struct came from [davecheney/manta](https://github.com/davecheney/manta)
