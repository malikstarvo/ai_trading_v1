#!/usr/bin/env bash
grep -oP 'https://[a-z-]+\.trycloudflare\.com' /home/ubuntu/ai_trading_v1/logs/cloudflared.log | tail -1
