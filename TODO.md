# TASKS

## Step 1: Simple REST client 
- [] Add tests to the REST client

## Step 2: Add a Write-Ahead Log (WAL) for Durability

**To prevent data loss, every write operation must be saved to disk before it's applied to the MemTable. This log file is called a Write-Ahead Log (WAL).

Goal: Make the KV store durable so it can survive a crash.

How-to:
1. Modify the Set and Delete methods. Before touching the in-memory map, open a log file (wal.log) in append-only mode.
2. Write the operation as a simple record. You can use Go's encoding/gob or JSON to serialize a LogEntry struct.

On startup, your NewKVStore() function will now first read the wal.log file from start to finish, replaying each entry into the MemTable to restore its state.

MVP #2: A durable, in-memory KV store. It's still limited by your machine's RAM, but if it crashes, you can restart it and all your data will be there.

### Step 3: Flush the MemTable to an SSTable

The MemTable can't grow forever. When it reaches a certain size (e.g., 10,000 keys), we need to flush it to a permanent file on disk called an SSTable (Sorted String Table).

Goal: Allow the database to store more data than can fit in the MemTable.

How-to:
1. Add a size check inside your Set method. When the MemTable hits the threshold, trigger a flush.
2. The Flush Process:
   1. Swap the full MemTable with a new, empty one. The old one is now "frozen." Incoming writes go to the new MemTable.
   2. Take all the key-value pairs from the frozen MemTable.
   3. Sort them alphabetically by key.

3. Write the sorted data to a new, uniquely named file on disk (e.g., sst_1.db). This is your first SSTable. The file format can be simple (e.g., key length, key, value length, value).

4. Once the SSTable is safely on disk, you can delete the old WAL file that was backing the frozen MemTable.

MVP #3: A KV store that offloads data from memory to sorted, immutable files on disk. Writes are still fast, but we can't read from the disk files yet.