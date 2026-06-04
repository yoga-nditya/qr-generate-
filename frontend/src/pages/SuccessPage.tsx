import React from 'react';
import { useSearchParams, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Check, Calendar, Receipt, User, Sparkles, ShieldCheck, QrCode } from 'lucide-react';

export const SuccessPage: React.FC = () => {
  const [searchParams] = useSearchParams();
  const navigate = useNavigate();

  const orderId  = searchParams.get('order_id') || 'SIM-DUMMY-ORDER';
  const txnId    = searchParams.get('txn_id')   || 'TXN-DUMMY-1234';
  const amount   = parseInt(searchParams.get('amount') || '10000', 10);

  const formatRupiah = (num: number) =>
    new Intl.NumberFormat('id-ID', {
      style: 'currency',
      currency: 'IDR',
      minimumFractionDigits: 0,
    }).format(num);

  const getFormattedDate = () =>
    new Intl.DateTimeFormat('id-ID', {
      dateStyle: 'medium',
      timeStyle: 'short',
    }).format(new Date());

  return (
    <div className="page-wrapper">
      <div className="page-container">
        <motion.div
          initial={{ opacity: 0, y: 18 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.45, ease: 'easeOut' }}
          style={{ width: '100%', display: 'flex', flexDirection: 'column', alignItems: 'center', gap: '14px' }}
        >
          {/* Brand Header */}
          <div className="app-header">
            <div className="app-logo-container">
              <QrCode size={22} />
            </div>
            <h1 className="app-title">QRIS Merchant</h1>
            <p className="app-subtitle">Bukti Pembayaran</p>
          </div>

          {/* Success Card */}
          <div className="payment-card">

            {/* Animated Checkmark */}
            <div className="receipt-header">
              <div className="success-checkmark-wrapper">
                <motion.div
                  initial={{ scale: 0.4, opacity: 0 }}
                  animate={{ scale: [1, 1.12, 1], opacity: 1 }}
                  transition={{ duration: 0.5, ease: 'easeOut' }}
                  style={{
                    position: 'absolute',
                    inset: 0,
                    backgroundColor: 'var(--success-light)',
                    borderRadius: '50%',
                  }}
                />
                <motion.div
                  initial={{ scale: 0 }}
                  animate={{ scale: 1 }}
                  transition={{ type: 'spring', stiffness: 280, damping: 18, delay: 0.1 }}
                  style={{ position: 'relative', display: 'flex', zIndex: 2 }}
                >
                  <Check size={30} strokeWidth={3} />
                </motion.div>
                <motion.div
                  initial={{ scale: 0, opacity: 0 }}
                  animate={{ scale: 1, opacity: 1 }}
                  transition={{ delay: 0.45 }}
                  className="success-sparkle"
                >
                  <Sparkles size={15} />
                </motion.div>
              </div>

              <h2 className="app-title" style={{ fontSize: '17px', marginTop: '6px' }}>
                Pembayaran Berhasil!
              </h2>
              <div className="receipt-status-pill">
                <span style={{ width: 6, height: 6, borderRadius: '50%', backgroundColor: '#10b981' }} />
                Settlement
              </div>
            </div>

            {/* Amount */}
            <div className="amount-container" style={{ marginBottom: '14px' }}>
              <p className="amount-label">Jumlah Terbayar</p>
              <h2 className="amount-value" style={{ fontSize: '22px' }}>{formatRupiah(amount)}</h2>
            </div>

            {/* Receipt Detail */}
            <div className="receipt-card">
              <div className="receipt-list">

                <div className="receipt-row">
                  <span className="receipt-label">
                    <Receipt size={13} />
                    ID Order
                  </span>
                  <span className="receipt-value mono">{orderId}</span>
                </div>

                <div className="receipt-row">
                  <span className="receipt-label">
                    <User size={13} />
                    Penerima
                  </span>
                  <span className="receipt-value">Toko Simulasi Mandiri</span>
                </div>

                <div className="receipt-row">
                  <span className="receipt-label">
                    <Calendar size={13} />
                    Waktu
                  </span>
                  <span className="receipt-value">{getFormattedDate()}</span>
                </div>

                <div className="receipt-divider" />

                <div className="receipt-row">
                  <span className="receipt-label">ID Transaksi</span>
                  <span className="receipt-value mono">{txnId}</span>
                </div>

                <div className="receipt-row">
                  <span className="receipt-label">Metode</span>
                  <span className="receipt-value">QRIS (Simulasi)</span>
                </div>
              </div>
            </div>

            {/* Test QR Lagi Button */}
            <button className="btn-action" onClick={() => navigate('/')}>
              <QrCode size={16} />
              <span>Test QR Lagi</span>
            </button>
          </div>

          {/* Footer */}
          <div className="page-footer">
            <ShieldCheck size={12} />
            <span>Sistem simulasi otomatis (tanpa Midtrans)</span>
          </div>
        </motion.div>
      </div>
    </div>
  );
};
