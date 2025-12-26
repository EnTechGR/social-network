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