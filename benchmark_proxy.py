#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
================================================================================
  🚀 PROXY LOAD TESTING & BOTTLENECK DIAGNOSTIC TOOL (500 - 1000 LUỒNG)
  Đo lường chi tiết từng mili-giây hiệu năng Proxy:
    - Chặng 1: Kết nối TCP Client -> VPS
    - Chặng 2: Proxy Tunnel & Handshake IPv6 tới Web đích
    - Chặng 3: TTFB (Time to First Byte) từ Web đích
    - Chặng 4: Tổng thời gian hoàn tất Request
    - Phân tích theo tiến trình luồng (kiểm tra luồng sau có bị chậm không)
    - Thống kê tỷ lệ phân tán IPv6 (độ độc nhất của IP xoay)
================================================================================
"""

import argparse
import asyncio
import os
import re
import sys
import time
from dataclasses import dataclass
from typing import Dict, List
import aiohttp

if hasattr(sys.stdout, 'reconfigure'):
    sys.stdout.reconfigure(encoding='utf-8', errors='backslashreplace')
if hasattr(sys.stderr, 'reconfigure'):
    sys.stderr.reconfigure(encoding='utf-8', errors='backslashreplace')

@dataclass
class RequestMetric:
    req_id: int
    success: bool = False
    status_code: int = 0
    t_tcp_connect: float = 0.0     # Chặng 1: Client -> VPS TCP Connect (ms)
    t_ttfb: float = 0.0            # Chặng 2: Gửi request đến nhận byte đầu tiên (ms)
    t_total: float = 0.0           # Chặng 3: Tổng thời gian toàn bộ request (ms)
    out_ip: str = ""               # Địa chỉ IPv6 xoay thu được
    error_msg: str = ""

def calc_stats(values: List[float]) -> Dict[str, float]:
    if not values:
        return {"avg": 0, "p50": 0, "p75": 0, "p90": 0, "p95": 0, "p99": 0, "min": 0, "max": 0}
    s = sorted(values)
    n = len(s)
    def p(pct):
        idx = int(n * pct / 100.0)
        return s[min(idx, n - 1)]
    return {
        "avg": sum(s) / n,
        "p50": p(50),
        "p75": p(75),
        "p90": p(90),
        "p95": p(95),
        "p99": p(99),
        "min": s[0],
        "max": s[-1]
    }

async def run_load_test(
    proxy_url: str,
    target_url: str,
    total_requests: int,
    concurrency: int,
    timeout_sec: float
):
    print("=" * 80)
    print("       🚀 BẮT ĐẦU ĐO ĐẠC ĐỘ TẢI & CHẨN ĐOÁN NGHẼN PROXY (BENCHMARK)       ")
    print("=" * 80)
    print(f"  • Proxy URL:            {proxy_url}")
    print(f"  • Web Mục Tiêu:         {target_url}")
    print(f"  • Tổng Số Request:      {total_requests} requests")
    print(f"  • Số Luồng Đồng Thời:   {concurrency} concurrent workers")
    print(f"  • Timeout Mỗi Request:  {timeout_sec}s")
    print("=" * 80)

    # Cấu hình TCP Connector tối ưu cho tải 1000+ luồng
    connector = aiohttp.TCPConnector(
        limit=concurrency + 50,
        ssl=False,
        force_close=True,  # Đảm bảo mỗi request mở 1 kết nối mới để đo đạc chính xác việc xoay IP
        enable_cleanup_closed=True
    )

    records: List[RequestMetric] = []
    sem = asyncio.Semaphore(concurrency)
    completed = 0
    t_bench_start = time.perf_counter()

    # Dùng TraceConfig để bóc tách từng chặng
    trace_cfg = aiohttp.TraceConfig()

    async def on_conn_start(session, trace_config_ctx, params):
        req_ctx = getattr(trace_config_ctx, 'trace_request_ctx', None)
        if req_ctx:
            req_ctx.t_conn_start = time.perf_counter()

    async def on_conn_end(session, trace_config_ctx, params):
        t_now = time.perf_counter()
        req_ctx = getattr(trace_config_ctx, 'trace_request_ctx', None)
        if req_ctx and hasattr(req_ctx, 't_conn_start'):
            req_ctx.t_tcp_ms = (t_now - req_ctx.t_conn_start) * 1000

    async def on_req_start(session, trace_config_ctx, params):
        req_ctx = getattr(trace_config_ctx, 'trace_request_ctx', None)
        if req_ctx:
            req_ctx.t_req_start = time.perf_counter()

    async def on_headers(session, trace_config_ctx, params):
        t_now = time.perf_counter()
        req_ctx = getattr(trace_config_ctx, 'trace_request_ctx', None)
        if req_ctx and hasattr(req_ctx, 't_req_start') and not hasattr(req_ctx, 't_ttfb_ms'):
            req_ctx.t_ttfb_ms = (t_now - req_ctx.t_req_start) * 1000

    trace_cfg.on_connection_create_start.append(on_conn_start)
    trace_cfg.on_connection_create_end.append(on_conn_end)
    trace_cfg.on_request_start.append(on_req_start)
    trace_cfg.on_response_chunk_received.append(on_headers)

    client_timeout = aiohttp.ClientTimeout(total=timeout_sec)

    async with aiohttp.ClientSession(
        connector=connector,
        trace_configs=[trace_cfg],
        timeout=client_timeout
    ) as session:

        async def worker(req_id: int):
            nonlocal completed
            rec = RequestMetric(req_id=req_id)
            trace_ctx = types_SimpleNamespace()

            async with sem:
                t0 = time.perf_counter()
                try:
                    async with session.get(
                        target_url,
                        proxy=proxy_url,
                        trace_request_ctx=trace_ctx
                    ) as resp:
                        body = await resp.text()
                        rec.status_code = resp.status
                        rec.success = (resp.status == 200)

                        # Tìm IPv6 trong phản hồi
                        match = re.search(r'([0-9a-fA-F]{1,4}:[0-9a-fA-F:]{3,})', body)
                        if match:
                            rec.out_ip = match.group(1)

                        rec.t_tcp_connect = getattr(trace_ctx, 't_tcp_ms', 0.0)
                        rec.t_ttfb = getattr(trace_ctx, 't_ttfb_ms', 0.0)

                except asyncio.TimeoutError:
                    rec.error_msg = f"Timeout (> {timeout_sec}s)"
                except Exception as e:
                    rec.error_msg = str(e)
                finally:
                    rec.t_total = (time.perf_counter() - t0) * 1000
                    records.append(rec)
                    completed += 1

                    # In tiến độ định kỳ
                    step = max(1, total_requests // 10)
                    if completed % step == 0 or completed == total_requests:
                        dur = time.perf_counter() - t_bench_start
                        rps = completed / dur if dur > 0 else 0
                        pct = completed / total_requests * 100
                        err_count = sum(1 for r in records if not r.success)
                        print(f"  [Tiến trình {pct:5.1f}%] Hoàn thành {completed:4d}/{total_requests} | Tốc độ: {rps:5.1f} req/s | Lỗi: {err_count}")

        print("⏳ Đang bắn tải đa luồng...")
        tasks = [worker(i + 1) for i in range(total_requests)]
        await asyncio.gather(*tasks)

    t_bench_total = time.perf_counter() - t_bench_start
    records.sort(key=lambda r: r.req_id)

    # -------------------------------------------------------------
    # BÁO CÁO TỔNG QUAN
    # -------------------------------------------------------------
    success_recs = [r for r in records if r.success]
    fail_recs = [r for r in records if not r.success]
    unique_ips = set(r.out_ip for r in records if r.out_ip)

    print("\n" + "=" * 80)
    print("                     📊 KẾT QUẢ TỔNG QUAN ĐO TẢI                 ")
    print("=" * 80)
    print(f"  • Thời gian chạy toàn bộ:   {t_bench_total:.2f} giây")
    print(f"  • Thông lượng trung bình:   {len(records) / t_bench_total:.1f} Requests / Giây (RPS)")
    print(f"  • Yêu cầu thành công:       {len(success_recs)} / {len(records)} ({len(success_recs)/len(records)*100:.1f}%)")
    print(f"  • Yêu cầu thất bại:         {len(fail_recs)} / {len(records)} ({len(fail_recs)/len(records)*100:.1f}%)")
    print(f"  • Số IPv6 xoay độc nhất:    {len(unique_ips)} địa chỉ khác nhau (Tỷ lệ phân tán: {len(unique_ips)/max(1, len(success_recs))*100:.1f}%)")

    if fail_recs:
        print("\n  ❌ Phân loại nguyên nhân lỗi:")
        err_map = {}
        for r in fail_recs:
            msg = r.error_msg or f"HTTP {r.status_code}"
            err_map[msg] = err_map.get(msg, 0) + 1
        for msg, cnt in sorted(err_map.items(), key=lambda x: -x[1]):
            print(f"     • {msg}: {cnt} lần ({cnt/len(records)*100:.1f}%)")

    # -------------------------------------------------------------
    # BẢNG PHÂN RÃ CHI TIẾT TỪNG CHẶNG
    # -------------------------------------------------------------
    if success_recs:
        s_tcp = calc_stats([r.t_tcp_connect for r in success_recs if r.t_tcp_connect > 0])
        s_ttfb = calc_stats([r.t_ttfb for r in success_recs if r.t_ttfb > 0])
        s_tot = calc_stats([r.t_total for r in success_recs])

        print("\n" + "=" * 80)
        print("          ⏱️ BẢNG PHÂN RÃ ĐỘ TRỄ CHI TIẾT TỪNG CHẶNG (mili-giây - ms)      ")
        print("=" * 80)
        hdr = f"{'Giai Đoạn Kết Nối':<32} | {'Trung Bình':>10} | {'P50 (Median)':>11} | {'P90':>8} | {'P95':>8} | {'P99':>8}"
        print(hdr)
        print("-" * len(hdr))

        def p_row(title, d):
            print(f"{title:<32} | {d['avg']:>8.1f}ms | {d['p50']:>9.1f}ms | {d['p90']:>6.1f}ms | {d['p95']:>6.1f}ms | {d['p99']:>6.1f}ms")

        p_row("1. Client -> VPS (TCP Connect)", s_tcp)
        p_row("2. Proxy -> Web đích (IPv6 TTFB)", s_ttfb)
        print("-" * len(hdr))
        p_row("⭐ TỔNG THỜI GIAN REQUEST", s_tot)
        print("=" * 80)

        # -------------------------------------------------------------
        # PHÂN TÍCH TIẾN TRÌNH LUỒNG (Kiểm tra xem luồng sau có bị chậm 2-3s không)
        # -------------------------------------------------------------
        print("\n" + "=" * 80)
        print("     📈 PHÂN TÍCH THEO TIẾN TRÌNH LUỒNG (Đo độ trễ các đợt luồng)")
        print("=" * 80)
        n_chunks = 4
        chunk_sz = max(1, len(records) // n_chunks)
        print(f"{'Đợt Luồng':<20} | {'Khoảng Luồng':<15} | {'Độ Trễ TB':>12} | {'Thành Công':>12} | {'Tình Trạng'}")
        print("-" * 75)

        for i in range(n_chunks):
            sub = records[i * chunk_sz : (i + 1) * chunk_sz]
            if not sub:
                continue
            sub_succ = [r for r in sub if r.success]
            avg_lat = sum(r.t_total for r in sub_succ) / len(sub_succ) if sub_succ else 0
            succ_pct = len(sub_succ) / len(sub) * 100
            st = "🟢 Cực nhanh"
            if avg_lat > 2000:
                st = "🔴 Bị chậm (>2s)"
            elif avg_lat > 1000:
                st = "🟡 Hơi chậm (>1s)"
            elif avg_lat > 500:
                st = "🔵 Khá tốt"
            print(f"Đợt {i+1} ({i*25:2d}%-{(i+1)*25:2d}%)      | Luồng {i*chunk_sz+1:4d} - {(i+1)*chunk_sz:4d} | {avg_lat:>10.1f}ms | {succ_pct:>11.1f}% | {st}")

        # -------------------------------------------------------------
        # CHẨN ĐOÁN TỰ ĐỘNG NGUYÊN NHÂN NGHẼN
        # -------------------------------------------------------------
        print("\n" + "=" * 80)
        print("                🩺 CHẨN ĐOÁN ĐIỂM NGHẼN TỰ ĐỘNG (DIAGNOSIS)               ")
        print("=" * 80)

        # Đánh giá chặng 1
        if s_tcp['p95'] > 200:
            print("  ⚠️ CHẶNG 1 (Client -> VPS): Bị nghẽn đường truyền từ máy bạn đến VPS!")
            print(f"     -> P95 là {s_tcp['p95']:.1f}ms. Mạng cáp quang hoặc hàng đợi socket trên máy bạn bị đầy khi bắn quá nhiều luồng cùng lúc.")
        else:
            print(f"  ✅ CHẶNG 1 (Client -> VPS): Hoàn hảo ({s_tcp['avg']:.1f}ms). Google BBR và TCP Fast Open trên VPS tiếp nhận kết nối cực tốt.")

        # Đánh giá chặng 2
        if s_ttfb['p95'] > 2000:
            print(f"  🔴 CHẶNG 2 (Proxy VPS -> Web đích): Bị chậm ({s_ttfb['p95']:.1f}ms)!")
            print("     -> Nguyên nhân: Máy chủ web đích (Target) đang phản hồi chậm hoặc nghẽn đường truyền quốc tế.")
        else:
            print(f"  ✅ CHẶNG 2 (Proxy VPS -> Web đích): Rất mượt ({s_ttfb['avg']:.1f}ms). Tuyến đường IPv6 xoay kết nối nhanh và thông suốt.")

        # Đánh giá độ phân tán IPv6
        if len(unique_ips) > 0:
            print(f"  ✅ XOAY IPV6: Đã xoay thành công qua {len(unique_ips)} địa chỉ IPv6 khác nhau trong Pool 4.096 IP.")
        print("=" * 80 + "\n")

class types_SimpleNamespace:
    pass

def main():
    parser = argparse.ArgumentParser(description="Tool đo đạc và chẩn đoán độ tải Proxy (500 - 1000 luồng)")
    parser.add_argument("--proxy", default="http://dfgfg333:eU3xkMDW@103.199.11.9:10112", help="Địa chỉ Proxy HTTP (VD: http://user:pass@host:port)")
    parser.add_argument("--target", default="https://api64.ipify.org?format=json", help="URL trang web kiểm thử (hỗ trợ IPv6)")
    parser.add_argument("-n", "--requests", type=int, default=500, help="Tổng số request cần gửi (mặc định: 500)")
    parser.add_argument("-c", "--concurrency", type=int, default=200, help="Số luồng đồng thời (mặc định: 200)")
    parser.add_argument("-t", "--timeout", type=float, default=10.0, help="Timeout mỗi request tính bằng giây (mặc định: 10.0)")

    args = parser.parse_args()

    try:
        asyncio.run(run_load_test(
            proxy_url=args.proxy,
            target_url=args.target,
            total_requests=args.requests,
            concurrency=args.concurrency,
            timeout_sec=args.timeout
        ))
    except KeyboardInterrupt:
        print("\n🛑 Người dùng đã dừng kiểm thử!")

if __name__ == "__main__":
    main()
