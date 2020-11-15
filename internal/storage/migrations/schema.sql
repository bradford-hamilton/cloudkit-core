-- Create table for storing VM data
CREATE TABLE IF NOT EXISTS vms (
  id SERIAL NOT NULL PRIMARY KEY,
  domain_id INT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Create table for storing VM memory snapshots --
CREATE TABLE IF NOT EXISTS measurements (
  id SERIAL NOT NULL PRIMARY KEY,
	time TIMESTAMPTZ NOT NULL,
	vm_id INT NOT NULL,
 	mem_usage DOUBLE PRECISION NOT NULL,
  CONSTRAINT fk_vm FOREIGN KEY(vm_id) REFERENCES vms(id)
);
