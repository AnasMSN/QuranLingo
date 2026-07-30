import { getDb } from './database';
import type { CourseTree, LessonDetail } from '../types/api';

export async function getCachedCourse(code: string): Promise<CourseTree | null> {
  const db = await getDb();
  const row = await db.getFirstAsync<{ payload: string }>(
    'SELECT payload FROM course_cache WHERE code = ?',
    [code],
  );
  return row ? (JSON.parse(row.payload) as CourseTree) : null;
}

export async function setCachedCourse(code: string, tree: CourseTree): Promise<void> {
  const db = await getDb();
  await db.runAsync(
    'INSERT OR REPLACE INTO course_cache (code, payload, updated_at) VALUES (?, ?, ?)',
    [code, JSON.stringify(tree), Date.now()],
  );
}

export async function getCachedLesson(id: string): Promise<LessonDetail | null> {
  const db = await getDb();
  const row = await db.getFirstAsync<{ payload: string }>(
    'SELECT payload FROM lesson_cache WHERE id = ?',
    [id],
  );
  return row ? (JSON.parse(row.payload) as LessonDetail) : null;
}

export async function setCachedLesson(id: string, lesson: LessonDetail): Promise<void> {
  const db = await getDb();
  await db.runAsync(
    'INSERT OR REPLACE INTO lesson_cache (id, payload, updated_at) VALUES (?, ?, ?)',
    [id, JSON.stringify(lesson), Date.now()],
  );
}
