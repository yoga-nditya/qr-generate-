import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { QRCodeSVG } from 'qrcode.react';
import { motion } from 'framer-motion';
import { Clock, Loader2, Scan } from 'lucide-react';

interface OrderData {
  order_id: string;
  amount: number;
  qr_string: string;
  status: string;
  created_at: string;
}

export const PaymentPage: React.FC = () => {
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [order, setOrder] = useState<OrderData | null>(null);
  const [timeLeft, setTimeLeft] = useState<number>(300);
  const navigate = useNavigate();

  // Create order on mount
  useEffect(() => {
    const createOrder = async () => {
      try {
        setLoading(true);
        const res = await fetch('/api/simulate/create', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
        });
        if (!res.ok) throw new Error('Gagal membuat kode pembayaran QRIS');
        const json = await res.json();
        if (json.success && json.data) {
          setOrder(json.data);
        } else {
          throw new Error(json.message || 'Gagal membuat kode pembayaran');
        }
      } catch (err: any) {
        setError(err.message || 'Koneksi internet bermasalah');
      } finally {
        setLoading(false);
      }
    };
    createOrder();
  }, []);

  // Countdown timer
  useEffect(() => {
    if (!order || timeLeft <= 0) return;
    const timer = setInterval(() => setTimeLeft((prev) => prev - 1), 1000);
    return () => clearInterval(timer);
  }, [order, timeLeft]);

  // WebSocket: listen for payment_update → auto navigate
  useEffect(() => {
    if (!order) return;

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//${window.location.host}/ws/${order.order_id}`;
    let socket: WebSocket;

    try {
      socket = new WebSocket(wsUrl);

      socket.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          if (data.type === 'payment_update' && data.status === 'success') {
            const txnId = data.data?.transaction_id || 'TXN-WS-DUMMY';
            navigate(
              `/success?order_id=${order.order_id}&txn_id=${txnId}&amount=${order.amount}`
            );
          }
        } catch (e) {
          console.error('Error parsing WebSocket message:', e);
        }
      };

      socket.onerror = (err) => console.error('WebSocket error:', err);
    } catch (e) {
      console.error('Failed to create WebSocket:', e);
    }

    return () => {
      if (socket) socket.close();
    };
  }, [order, navigate]);

  const formatRupiah = (amount: number) =>
    new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(amount);

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  /* ── Loading ── */
  if (loading) {
    return (
      <div className="page-wrapper">
        <div className="page-container">
          <div className="payment-card">
            <div className="loading-state">
              <motion.div
                animate={{ rotate: 360 }}
                transition={{ repeat: Infinity, duration: 1, ease: 'linear' }}
                style={{ color: 'var(--primary-color)' }}
              >
                <Loader2 size={36} />
              </motion.div>
              <p>Menyiapkan kode QRIS...</p>
            </div>
          </div>
        </div>
      </div>
    );
  }

  /* ── Error ── */
  if (error || !order) {
    return (
      <div className="page-wrapper">
        <div className="page-container">
          <div className="payment-card">
            <div className="error-state">
              <h2 style={{ color: '#ef4444', fontWeight: 800, fontSize: '16px' }}>
                Gagal Memuat
              </h2>
              <p style={{ color: 'var(--text-secondary)', fontSize: '13px', marginTop: '4px' }}>
                {error || 'Gagal memproses data order.'}
              </p>
              <button className="btn-retry" onClick={() => window.location.reload()}>
                Coba Lagi
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  /* ── Main ── */
  return (
    <div className="page-wrapper">
      <div className="page-container">
        <motion.div
          initial={{ opacity: 0, y: 18 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.45, ease: 'easeOut' }}
          style={{ width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '14px' }}
        >

          {/* Payment Card */}
          <div className="payment-card">

            {/* Merchant Info */}
            <div className="merchant-info">
              <p className="merchant-label">Nama Merchant</p>
              <h3 className="merchant-name">TOKO SIMULASI MANDIRI</h3>
              <p className="merchant-nmid">NMID: ID1020304050607</p>
            </div>

            {/* Amount */}
            <div className="amount-container">
              <p className="amount-label">Total Pembayaran</p>
              <h2 className="amount-value">{formatRupiah(order.amount)}</h2>
            </div>

            {/* QR Code — scan only, no click needed */}
            <div className="qr-section">
              <motion.div
                className="qr-frame"
                initial={{ scale: 0.95, opacity: 0 }}
                animate={{ scale: 1, opacity: 1 }}
                transition={{ delay: 0.15, duration: 0.4, ease: 'easeOut' }}
              >
                <QRCodeSVG
                  value={order.qr_string}
                  size={180}
                  level="M"
                  includeMargin={true}
                />
              </motion.div>

              <div className="qr-scan-hint">
                <Scan size={13} />
                <span>Scan dengan aplikasi e-wallet / mobile banking</span>
              </div>
            </div>

            {/* Status Panel */}
            <div className="status-panel">
              <div className="status-item">
                <Clock size={12} style={{ color: 'var(--text-muted)' }} />
                <span>
                  Sisa Waktu:{' '}
                  <span className="time-countdown">{formatTime(timeLeft)}</span>
                </span>
              </div>
              <div className="status-divider" />
              <div className="status-item">
                <span className="status-indicator" />
                <span>Menunggu pembayaran</span>
              </div>
            </div>
          </div>

        </motion.div>
      </div>
    </div>
  );
};
