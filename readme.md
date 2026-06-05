Cara Jalanin Proyek
ssh -D 1080 ubuntu@43.159.56.168 -N

# 1. Start database
docker-compose up -d postgres

# 2. Start collector (isi candles, OI, funding, L/S, liquidation)
$env:COLLECTOR_PROXY_URL="socks5://127.0.0.1:1080"; $env:COLLECTOR_TESTNET="false"; go run ./cmd/trader/

go run ./cmd/trader/
<!-- $env:COLLECTOR_INSECURE_SKIP_VERIFY="true"; go run ./cmd/trader/ -->

# Tunggu beberapa menit sampai data terkumpul...

# 3. Backfill features + labels
go run ./cmd/feature-backfill/

# 4. Baru edge study
go run ./cmd/edge-study/

# Hasil: research_artifacts/edge_report.html (buka di browser)
Catatan: Testnet Bybit gratis — tidak perlu API key untuk public endpoints (kline, open interest, funding rate, liquidation, L/S ratio). Cuma butuh API key kalau mau trading (private endpoints).