# Project Outline

## Step 1: The MemTable (Your Current In-Memory Store)

This is your starting point. You have a struct containing a Go map and a sync.RWMutex to handle concurrent access. From now on, we'll call this in-memory map the MemTable, because it's a table of data stored in memory.

Goal: A thread-safe, in-memory key-value store.

Functionality: Set(key, value), Get(key), Delete(key).

MVP #1: A simple KV store that works perfectly but loses all its data when the program stops. This is the foundation we'll build upon.

// Your starting point
type KVStore struct {
    mu   sync.RWMutex
    memtable map[string]string // This is our MemTable
}

## Step 2: Add a Write-Ahead Log (WAL) for Durability

**To prevent data loss, every write operation must be saved to disk before it's applied to the MemTable. This log file is called a Write-Ahead Log (WAL).

Goal: Make the KV store durable so it can survive a crash.

How-to:
1. Modify the Set and Delete methods. Before touching the in-memory map, open a log file (wal.log) in append-only mode.
2. Write the operation as a simple record. You can use Go's encoding/gob or JSON to serialize a LogEntry struct.

On startup, your NewKVStore() function will now first read the wal.log file from start to finish, replaying each entry into the MemTable to restore its state.

MVP #2: A durable, in-memory KV store. It's still limited by your machine's RAM, but if it crashes, you can restart it and all your data will be there.

## Step 3: Flush the MemTable to an SSTable

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

## Step 4: Unify Reads from MemTable and SSTables

Now that data lives in two places (the active MemTable and one or more SSTable files on disk), your Get function needs to get smarter.

Goal: Read a key's value, regardless of where it's stored.

How-to:
1. Modify the Get(key) method to search in a specific order, from newest data to oldest.
2. Lookup Order:
   1. First, check the MemTable. If the key is there, return the value.
   2. If not found, search the SSTable files. You'll need to check them from newest to oldest (e.g., sst_2.db, then sst_1.db).
   3. Searching SSTables: For your MVP, you can just read through each SSTable file from the beginning to find the key. Since the file is sorted, you'll know if the key doesn't exist once you pass where it should be.

3. Handling Deletes: When you Delete a key, write a special "tombstone" marker into the MemTable and then into the SSTable. When reading, if Get finds a tombstone, it stops immediately and returns "not found," even if an older SSTable has the key.

MVP #4: A fully functional KV store. You can write data, and it will be flushed to disk. You can read data, and it will be found whether it's in memory or on disk.

## Step 5: Implement Compaction

Over time, you'll accumulate many SSTable files. Searching through all of them for a key is slow. Compaction is a background process that merges SSTables to reduce their number and clean up old data.

Goal: Keep read performance high and reclaim disk space from deleted/overwritten data.

How-to:

1. Create a background goroutine that runs periodically.

2. Compaction Logic (Simplified):
   1. The goroutine selects several SSTables (for a simple MVP, it could just select all of them).
   2. It reads all the selected files simultaneously, using a k-way merge algorithm (like the one in merge sort) to produce a single stream of key-value pairs that is fully sorted.
   3. As it merges, it applies logic: if it sees the same key multiple times, it only keeps the newest version. If it sees a tombstone, it discards the tombstone and any older versions of that key.
      1. The clean, merged output is written to a new SSTable file.
      2. Once the new file is written successfully, the old input files are deleted.

MVP #5: A complete, working LSM-tree KV store! It's durable, can handle large amounts of data, and automatically maintains itself to keep reads fast.