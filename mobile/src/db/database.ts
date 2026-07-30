import * as SQLite from 'expo-sqlite';

let dbPromise: Promise<SQLite.SQLiteDatabase> | null = null;

// Local cache of lesson/course content and a queue of not-yet-synced lesson
// submissions — the offline-first layer learners rely on when practicing
// without signal. See src/db/courseCache.ts and src/db/submissionQueue.ts.
export function getDb(): Promise<SQLite.SQLiteDatabase> {
  if (!dbPromise) {
    dbPromise = SQLite.openDatabaseAsync('quranlingo.db').then(async (db) => {
      await db.execAsync(`
        PRAGMA journal_mode = WAL;
        CREATE TABLE IF NOT EXISTS course_cache (
          code TEXT PRIMARY KEY NOT NULL,
          payload TEXT NOT NULL,
          updated_at INTEGER NOT NULL
        );
        CREATE TABLE IF NOT EXISTS lesson_cache (
          id TEXT PRIMARY KEY NOT NULL,
          payload TEXT NOT NULL,
          updated_at INTEGER NOT NULL
        );
        CREATE TABLE IF NOT EXISTS pending_submissions (
          id TEXT PRIMARY KEY NOT NULL,
          lesson_id TEXT NOT NULL,
          lesson_title TEXT NOT NULL,
          idempotency_key TEXT NOT NULL,
          answers_json TEXT NOT NULL,
          created_at INTEGER NOT NULL
        );
      `);
      return db;
    });
  }
  return dbPromise;
}
