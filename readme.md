Cara Jalanin Proyek
# 1. Start database
docker-compose up -d postgres

# 2. Start collector (isi candles, OI, funding, L/S, liquidation)
go run ./cmd/trader/

# Tunggu beberapa menit sampai data terkumpul...

# 3. Backfill features + labels
go run cmd/feature-backfill/

# 4. Baru edge study
go run cmd/edge-study/

# Hasil: research_artifacts/edge_report.html (buka di browser)
Catatan: Testnet Bybit gratis — tidak perlu API key untuk public endpoints (kline, open interest, funding rate, liquidation, L/S ratio). Cuma butuh API key kalau mau trading (private endpoints).