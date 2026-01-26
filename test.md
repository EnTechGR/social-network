# Complete profiling suite
go test ./handlers \
  -bench=BenchmarkAuthHandler_Register \
  -benchmem \
  -benchtime=10s \
  -cpuprofile=cpu.prof \
  -memprofile=mem.prof

# View CPU in browser
go tool pprof -http=:8080 cpu.prof

# Or view in terminal
go tool pprof -top cpu.prof

go tool pprof -top cpu.prof


curl -X POST http://localhost:8080/api/v1/posts/create \
  -H "X-CSRF-Token: LbbbDcANt3sVv5joaKOzOtdITcd2JMje2NlraytIugs=" \
  -b alice1.cookies \
  -F "title=My Post with Images" \
  -F "content=Check out these amazing pictures" \
  -F "image=@/home/entech/win_documents/social-network/3551739.jpg" \
  -F "image=@/home/entech/win_documents/social-network/3551739.jpg"

