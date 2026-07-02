-- SIGMA Bisyaroh - Payroll Management System
-- PostgreSQL DDL (Supabase-compatible)

-- Enum types
CREATE TYPE employee_category AS ENUM ('BUK', 'S2_GEL1', 'S2_GEL2', 'REGULER', 'KSU');
CREATE TYPE activity_type AS ENUM ('ADDITION', 'DEDUCTION');
CREATE TYPE payroll_status AS ENUM ('DRAFT', 'CONFIRMED', 'PAID');

-- Employees table
CREATE TABLE employees (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    nip VARCHAR(50) UNIQUE,
    category employee_category NOT NULL,
    role VARCHAR(100),
    fixed_salary NUMERIC(19,4) DEFAULT 0,
    hourly_rate NUMERIC(19,4) DEFAULT 0,
    target_incentive NUMERIC(19,4) DEFAULT 0,
    structural_allowance NUMERIC(19,4) DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Master Activities lookup table
CREATE TABLE master_activities (
    id UUID PRIMARY KEY,
    activity_name VARCHAR(255) NOT NULL,
    code VARCHAR(50) UNIQUE NOT NULL,
    default_rate NUMERIC(19,4) DEFAULT 0,
    type activity_type NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Payroll Transactions (per employee per period)
CREATE TABLE payroll_transactions (
    id UUID PRIMARY KEY,
    employee_id UUID NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    month INTEGER NOT NULL CHECK (month BETWEEN 1 AND 12),
    year INTEGER NOT NULL CHECK (year BETWEEN 2020 AND 2100),
    gross_income NUMERIC(19,4) DEFAULT 0,
    total_deductions NUMERIC(19,4) DEFAULT 0,
    take_home_pay NUMERIC(19,4) DEFAULT 0,
    status payroll_status DEFAULT 'DRAFT',
    notes TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE(employee_id, month, year)
);

-- Payroll Details (line items)
CREATE TABLE payroll_details (
    id UUID PRIMARY KEY,
    payroll_transaction_id UUID NOT NULL REFERENCES payroll_transactions(id) ON DELETE CASCADE,
    activity_id UUID REFERENCES master_activities(id) ON DELETE SET NULL,
    quantity NUMERIC(19,4) DEFAULT 0,
    rate NUMERIC(19,4) DEFAULT 0,
    total_amount NUMERIC(19,4) DEFAULT 0,
    type activity_type NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_payroll_transactions_employee ON payroll_transactions(employee_id);
CREATE INDEX idx_payroll_transactions_period ON payroll_transactions(month, year);
CREATE INDEX idx_payroll_details_transaction ON payroll_details(payroll_transaction_id);
CREATE INDEX idx_employees_category ON employees(category);
CREATE INDEX idx_employees_active ON employees(is_active) WHERE is_active = true;

-- Seed: Master Activities
INSERT INTO master_activities (id, activity_name, code, default_rate, type, description) VALUES
    ('a0000001-0000-0000-0000-000000000001', 'Jam Mengajar', 'JAM_AJAR', 0, 'ADDITION', 'Perhitungan per jam mengajar'),
    ('a0000001-0000-0000-0000-000000000002', 'Lembur', 'LEMBUR', 0, 'ADDITION', 'Jam lembur di luar jadwal'),
    ('a0000001-0000-0000-0000-000000000003', 'Rapat', 'RAPAT', 0, 'ADDITION', 'Honor kehadiran rapat'),
    ('a0000001-0000-0000-0000-000000000004', 'Jaga Ujian', 'JAGA_UJIAN', 0, 'ADDITION', 'Honor pengawas ujian'),
    ('a0000001-0000-0000-0000-000000000005', 'Juri Lomba', 'JURI_LOMBA', 0, 'ADDITION', 'Honor juri perlombaan'),
    ('a0000001-0000-0000-0000-000000000006', 'Tauzi Fushul', 'TAUZI', 0, 'ADDITION', 'Honor pembagian kelas'),
    ('a0000001-0000-0000-0000-000000000007', 'Telat per Menit', 'TELAT', 1000, 'DEDUCTION', 'Potongan keterlambatan Rp1.000/menit'),
    ('a0000001-0000-0000-0000-000000000008', 'Potongan Absensi', 'ABSENSI', 0, 'DEDUCTION', 'Potongan tidak hadir tanpa keterangan');
