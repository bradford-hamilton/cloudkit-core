-- Create table for storing VM memory snapshots --
CREATE TABLE IF NOT EXISTS vms (
  id SERIAL NOT NULL PRIMARY KEY,
  domain_id INT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS measurements (
	time TIMESTAMPTZ NOT NULL,
	vm_id INT NOT NULL,
 	mem_usage DOUBLE PRECISION NOT NULL,
  CONSTRAINT fk_vm FOREIGN KEY(vm_id) REFERENCES vms(id)
);
