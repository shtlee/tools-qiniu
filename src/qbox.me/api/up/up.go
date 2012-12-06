package up

import (
    /*"crypto/hmac"
    "crypto/sha1"
    "encoding/json"*/
    "qbox.me/rpc"
    "net/http"
    "io"
    "fmt"
    "encoding/base64"
    "errors"
    . "qbox.me/api"
    "strconv"
    "hash/crc32"
    "bytes"
    "sync"
)

const (
    InvalidCtx = 701
    UP_HOST    = "http://up.qbox.me"
    BLOCK_BITS = 22
    PUT_CHUNK_SIZE = 256
    PUT_RETRY_TIMES = 3
)

// ----------------------------------------------------------
type UpService struct {
    *Config
    Conn rpc.Client
}


func New(c *Config, t http.RoundTripper) (s *UpService, err error) {
    if c == nil {
        err = errors.New("Must have a config file")
        return
    }
    if t == nil {
        t = http.DefaultTransport
    }
    client := &http.Client{Transport: t}
    s = &UpService{c, rpc.Client{c, client}}
    return
}

type PutRet struct {
    Ctx      string `json:"ctx"`
    Checksum string `json:"checksum"`
    Crc32    uint32 `json:"crc32"`
    Offset   uint32 `json:"offset"`
    Host     string `json:"host"`
}

func mkBlock(c rpc.Client, blockSize int, body io.Reader, bodyLength int) (ret PutRet, code int, err error) {
    code, err = c.CallWithBinaryEx(
        &ret, UP_HOST+"/mkblk/"+strconv.Itoa(blockSize), "application/octet-stream", body, bodyLength)
    return
}

func blockPut(c rpc.Client, ctx string, offset int, body io.Reader, bodyLength int) (ret PutRet, code int, err error) {
    code, err = c.CallWithBinaryEx(
        &ret, UP_HOST+"/bput/"+ctx+"/"+strconv.Itoa(offset), "application/octet-stream", body, bodyLength)
    return
}


type BlockProgress struct {
    Ctx      string
    Offset   int
    RestSize int
    Err      error
}

func blockFirstPut(prog *BlockProgress) bool {
    return prog.Ctx == ""
}

func getBodyLength(chunkSize, blkSize int) int {
    var bodyLength int
    if chunkSize < blkSize {
        bodyLength = chunkSize
    } else {
        bodyLength = blkSize
    }
    return bodyLength
}

func ResumableBlockput(
    c rpc.Client, f io.ReaderAt, blockIdx int, blkSize, chunkSize, retryTimes int,
    prog *BlockProgress, notify func(blockIdx int, prog *BlockProgress)) (ret PutRet, code int, err error) {

    offbase := int64(blockIdx) << BLOCK_BITS
    h := crc32.NewIEEE()

    var bodyLength int

    // The block never be uploaded.
    if blockFirstPut(prog){ 
        bodyLength = getBodyLength(chunkSize, blkSize)
        body1 := io.NewSectionReader(f, offbase, int64(bodyLength))
        body := io.TeeReader(body1, h)

        ret, code, err = mkBlock(c, blkSize, body, bodyLength)
        if err != nil {
            fmt.Println(" |- ResumaleBlockPut : mkblock failed : ", err)
            return
        }

        if ret.Crc32 != h.Sum32() {
            fmt.Println("ResumableBlockput: invalid checksum")
            return
        }

        prog.Ctx = ret.Ctx
        prog.Offset = bodyLength
        prog.RestSize = blkSize - bodyLength

        notify(blockIdx, prog)

    } else if prog.Offset+prog.RestSize != blkSize {
        code, err = 400, errors.New("Invalid args when doing ResumableBlockPut.")
        return
    }

    for prog.RestSize > 0 {
        if chunkSize < prog.RestSize {
            bodyLength = chunkSize
        } else {
            bodyLength = prog.RestSize
        }

        retry := retryTimes
    lzRetry:
        body1 := io.NewSectionReader(f, offbase+int64(prog.Offset), int64(bodyLength))
        h.Reset()
        body := io.TeeReader(body1, h)
        ret, code, err = blockPut(c, prog.Ctx, prog.Offset, body, bodyLength)

        // put successfully, but need more check should be done.
        if err == nil { 
            if ret.Crc32 == h.Sum32() {
                prog.Ctx = ret.Ctx
                prog.Offset += bodyLength
                prog.RestSize -= bodyLength
                notify(blockIdx, prog)
                continue
            } else {
                fmt.Println("ResumableBlockPut invalied checksum : ", offbase, prog.Offset, body)
            }
        } else {
            if code == InvalidCtx {
                fmt.Println("Invalid Context 701!")
                prog.Ctx = ""
                notify(blockIdx, prog)
                break
            }
        }

        if retry > 0 {
            retry--
            goto lzRetry
        }

        break
    }

    return
} 


func Mkfile(
    c rpc.Client, ret interface{}, cmd, entry string,
    fsize int64, params, callbackParams string, checksums []string) (code int, err error) {
    if callbackParams != "" {
        params += "/params/" + rpc.EncodeURI(callbackParams)
    }

    n := len(checksums)
    body := make([]byte, 20*n)
    for i, checksum := range checksums {
        ret, err2 := base64.URLEncoding.DecodeString(checksum)
        if err2 != nil {
            code, err = 400, errors.New("mkfile error")
            return
        }
        copy(body[i*20:], ret)
    }
    code, err = c.CallWithBinaryEx(
        ret, UP_HOST+cmd+rpc.EncodeURI(entry)+"/fsize/"+strconv.FormatInt(fsize, 10)+params,
        "application/octet-stream", bytes.NewReader(body), len(body))
    return
}

// a helper
func BlockCount(fsize int64) int {
    blockMask := int64((1 << BLOCK_BITS) - 1)
    return int((fsize + blockMask) >> BLOCK_BITS)
}

type Service struct {
    Tasks chan func()
}

func NewService(taskQsize, threadSize int) Service {
    tasks := make(chan func(), taskQsize)
    for i := 0; i < threadSize; i++ {
        go worker(tasks)
    }
    return Service{tasks}
}

func worker(tasks chan func()) {
    for {
        task := <-tasks
        task()
    }
}

func (r Service) Put(
    c rpc.Client, f io.ReaderAt, fsize int64, checksums []string, progs []BlockProgress,
    blockNotify func(blockIdx int, checksum string),
    chunkNotify func(blockIdx int, prog *BlockProgress)) (code int, err error) {

    blockCnt := BlockCount(fsize)
    if len(checksums) != blockCnt || len(progs) != blockCnt {
        code, err = 400, errors.New("up.Service.Put")
        return
    }

    var wg sync.WaitGroup
    wg.Add(blockCnt)
    last := blockCnt - 1
    blockSize := 1 << BLOCK_BITS
    var failed bool
    for i := 0; i < blockCnt; i++ {
        if checksums[i] == "" {
            blockIdx := i
            blockSize1 := blockSize
            if i == last {
                offbase := int64(blockIdx) << BLOCK_BITS
                blockSize1 = int(fsize - offbase)
            }
            task := func() {
                defer wg.Done()
                ret, code, err2 := ResumableBlockput(
                    c, f, blockIdx, blockSize1, PUT_CHUNK_SIZE, PUT_RETRY_TIMES, &progs[blockIdx], chunkNotify)
                if err2 != nil {
                    fmt.Println("ResumableBockPut", blockIdx, "failed", code, err2)
                    failed = true
                } else {
                    checksums[blockIdx] = ret.Checksum
                    blockNotify(blockIdx, ret.Checksum)
                }
                progs[blockIdx].Err = err2
            }
            r.Tasks <- task
        } else {
            wg.Done()
        }
    }

    wg.Wait()
    if failed {
        code, err = 201, errors.New("Function fails")
    } else {
        code = 200
    }
    return
}
