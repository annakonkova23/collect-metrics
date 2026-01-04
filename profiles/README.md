agent

File: main.exe
Build ID: C:\Users\AnnaK\AppData\Local\go-build\d7\d7b4c43a8aea00774cf80304040231cb53a032e65de107d9b5ab6ed273913dd4-d\main.exe2026-01-04 09:59:34.0548172 +0300 MSK
Type: inuse_space
Time: 2026-01-04 08:46:52 MSK
Duration: 60.01s, Total samples = 4403.16kB
Showing nodes accounting for -5372.16kB, 122.01% of 4403.16kB total
      flat  flat%   sum%        cum   cum%
-2707.76kB 61.50% 61.50% -4345.93kB 98.70%  compress/flate.NewWriter (inline)
-1638.17kB 37.20% 98.70% -1638.17kB 37.20%  compress/flate.(*compressor).initDeflate (inline)
 -513.12kB 11.65% 110.35%  -513.12kB 11.65%  compress/flate.(*huffmanEncoder).generate
    -513kB 11.65% 122.00%     -513kB 11.65%  runtime.allocm
 -512.09kB 11.63% 133.63% -4859.14kB 110.36%  github.com/annakonkova23/collect-metrics/internal/service.(*Sender).SendRequest
 -512.05kB 11.63% 145.26%  -512.05kB 11.63%  github.com/annakonkova23/collect-metrics/internal/service.(*MetricsUpdater).UpdateRuntimeMetric
  512.02kB 11.63% 133.64%   512.02kB 11.63%  container/list.(*List).insertValue (inline)
  512.01kB 11.63% 122.01% -4347.05kB 98.73%  github.com/annakonkova23/collect-metrics/internal/agent.(*Client).PostWithBody
         0     0% 122.01%  -513.12kB 11.65%  compress/flate.(*Writer).Close (inline)
         0     0% 122.01%  -513.12kB 11.65%  compress/flate.(*compressor).close
         0     0% 122.01%  -513.12kB 11.65%  compress/flate.(*compressor).deflate
         0     0% 122.01% -1638.17kB 37.20%  compress/flate.(*compressor).init
         0     0% 122.01%  -513.12kB 11.65%  compress/flate.(*compressor).writeBlock
         0     0% 122.01%  -513.12kB 11.65%  compress/flate.(*huffmanBitWriter).indexTokens
         0     0% 122.01%  -513.12kB 11.65%  compress/flate.(*huffmanBitWriter).writeBlock
         0     0% 122.01%  -513.12kB 11.65%  compress/gzip.(*Writer).Close
         0     0% 122.01% -4345.93kB 98.70%  compress/gzip.(*Writer).Write
         0     0% 122.01%   512.02kB 11.63%  container/list.(*List).PushFront (inline)
         0     0% 122.01%   512.02kB 11.63%  net/http.(*Transport).tryPutIdleConn
         0     0% 122.01%   512.02kB 11.63%  net/http.(*connLRU).add
         0     0% 122.01%   512.02kB 11.63%  net/http.(*persistConn).readLoop
         0     0% 122.01%   512.02kB 11.63%  net/http.(*persistConn).readLoop.func2
         0     0% 122.01%     -513kB 11.65%  runtime.mcall
         0     0% 122.01%     -513kB 11.65%  runtime.newm
         0     0% 122.01%     -513kB 11.65%  runtime.park_m
         0     0% 122.01%     -513kB 11.65%  runtime.resetspinning
         0     0% 122.01%     -513kB 11.65%  runtime.schedule
         0     0% 122.01%     -513kB 11.65%  runtime.startm
         0     0% 122.01%     -513kB 11.65%  runtime.wakep

server

File: main.exe
Build ID: C:\Users\AnnaK\AppData\Local\Temp\go-build2586814664\b001\exe\main.exe2026-01-04 10:55:43.6178255 +0300 MSK
Type: inuse_space
Time: 2026-01-04 08:32:54 MSK
Duration: 60.02s, Total samples = 1414.63kB
Showing nodes accounting for 2906.88kB, 205.49% of 1414.63kB total
Dropped 1 node (cum <= 7.07kB)
      flat  flat%   sum%        cum   cum%
 1805.17kB 127.61% 127.61%  2902.86kB 205.20%  compress/flate.NewWriter (inline)
 1097.69kB 77.60% 205.20%  1097.69kB 77.60%  compress/flate.(*compressor).initDeflate (inline)
  516.01kB 36.48% 241.68%   516.01kB 36.48%  io.init.func1
 -513.50kB 36.30% 205.38%  -513.50kB 36.30%  sync.(*Pool).pinSlow
     513kB 36.26% 241.64%  1025.56kB 72.50%  runtime.allocm
  512.56kB 36.23% 277.88%   512.56kB 36.23%  runtime.makeProfStackFP (inline)
 -512.05kB 36.20% 241.68%  -512.05kB 36.20%  internal/profile.decodeString (inline)
 -512.01kB 36.19% 205.49%  -512.01kB 36.19%  net/textproto.MIMEHeader.Set (inline)
         0     0% 205.49%   516.01kB 36.48%  bufio.(*Writer).Flush
         0     0% 205.49%  1097.69kB 77.60%  compress/flate.(*compressor).init
         0     0% 205.49%  2902.86kB 205.20%  compress/gzip.(*Writer).Write
         0     0% 205.49%  -512.01kB 36.19%  github.com/annakonkova23/collect-metrics/internal/handler.(*Server).StartAndListen.WithCheckHash.func1.1
         0     0% 205.49%  -512.01kB 36.19%  github.com/annakonkova23/collect-metrics/internal/handler.(*Server).updateSeveralJSONHandler
         0     0% 205.49%  1488.26kB 105.21%  github.com/annakonkova23/collect-metrics/internal/handler/middleware.WithCompress.func1
         0     0% 205.49%  1488.26kB 105.21%  github.com/annakonkova23/collect-metrics/internal/handler/middleware.WithLogging.func1
         0     0% 205.49%  2000.27kB 141.40%  github.com/annakonkova23/collect-metrics/internal/handler/middleware.codeResponse
         0     0% 205.49%  2000.27kB 141.40%  github.com/annakonkova23/collect-metrics/internal/handler/middleware.compessResponse
         0     0% 205.49%   974.76kB 68.91%  github.com/go-chi/chi/v5.(*Mux).ServeHTTP
         0     0% 205.49%  -512.01kB 36.19%  github.com/go-chi/chi/v5.(*Mux).routeHTTP
         0     0% 205.49%  -512.05kB 36.20%  internal/profile.Parse
         0     0% 205.49%  -512.05kB 36.20%  internal/profile.decodeMessage
         0     0% 205.49%  -512.05kB 36.20%  internal/profile.decodeStrings (inline)
         0     0% 205.49%  -512.05kB 36.20%  internal/profile.init.func6
         0     0% 205.49%  -512.05kB 36.20%  internal/profile.parseUncompressed
         0     0% 205.49%  -512.05kB 36.20%  internal/profile.unmarshal (inline)
         0     0% 205.49%   516.01kB 36.48%  io.Copy (inline)
         0     0% 205.49%   516.01kB 36.48%  io.CopyN
         0     0% 205.49%   516.01kB 36.48%  io.copyBuffer
         0     0% 205.49%   516.01kB 36.48%  io.discard.ReadFrom
         0     0% 205.49%   390.54kB 27.61%  net/http.(*ServeMux).ServeHTTP
         0     0% 205.49%   516.01kB 36.48%  net/http.(*chunkWriter).Write
         0     0% 205.49%   516.01kB 36.48%  net/http.(*chunkWriter).writeHeader
         0     0% 205.49%  1881.31kB 132.99%  net/http.(*conn).serve
         0     0% 205.49%   516.01kB 36.48%  net/http.(*response).finishRequest
         0     0% 205.49%  1878.80kB 132.81%  net/http.HandlerFunc.ServeHTTP
         0     0% 205.49%  -512.01kB 36.19%  net/http.Header.Set (inline)
         0     0% 205.49%  1365.30kB 96.51%  net/http.serverHandler.ServeHTTP
         0     0% 205.49%   390.54kB 27.61%  net/http/pprof.Index
         0     0% 205.49%   390.54kB 27.61%  net/http/pprof.collectProfile
         0     0% 205.49%   390.54kB 27.61%  net/http/pprof.handler.ServeHTTP
         0     0% 205.49%   390.54kB 27.61%  net/http/pprof.handler.serveDeltaProfile
         0     0% 205.49%   512.56kB 36.23%  runtime.mProfStackInit (inline)
         0     0% 205.49%   512.56kB 36.23%  runtime.mcall
         0     0% 205.49%   512.56kB 36.23%  runtime.mcommoninit
         0     0% 205.49%      513kB 36.26%  runtime.mstart
         0     0% 205.49%      513kB 36.26%  runtime.mstart0
         0     0% 205.49%      513kB 36.26%  runtime.mstart1
         0     0% 205.49%  1025.56kB 72.50%  runtime.newm
         0     0% 205.49%   512.56kB 36.23%  runtime.park_m
         0     0% 205.49%  1025.56kB 72.50%  runtime.resetspinning
         0     0% 205.49%  1025.56kB 72.50%  runtime.schedule
         0     0% 205.49%  1025.56kB 72.50%  runtime.startm
         0     0% 205.49%  1025.56kB 72.50%  runtime.wakep
         0     0% 205.49%   902.59kB 63.80%  runtime/pprof.(*Profile).WriteTo
         0     0% 205.49%   902.59kB 63.80%  runtime/pprof.(*profileBuilder).build
         0     0% 205.49%   902.59kB 63.80%  runtime/pprof.writeHeap
         0     0% 205.49%   902.59kB 63.80%  runtime/pprof.writeHeapInternal
         0     0% 205.49%   902.59kB 63.80%  runtime/pprof.writeHeapProto
         0     0% 205.49%  -513.50kB 36.30%  sync.(*Pool).pin