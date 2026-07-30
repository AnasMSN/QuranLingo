import { getDb } from './database';
import type { AnswerInput } from '../types/api';

export interface PendingSubmission {
  id: string;
  lessonId: string;
  lessonTitle: string;
  idempotencyKey: string;
  answers: AnswerInput[];
  createdAt: number;
}

interface PendingRow {
  id: string;
  lesson_id: string;
  lesson_title: string;
  idempotency_key: string;
  answers_json: string;
  created_at: number;
}

function fromRow(row: PendingRow): PendingSubmission {
  return {
    id: row.id,
    lessonId: row.lesson_id,
    lessonTitle: row.lesson_title,
    idempotencyKey: row.idempotency_key,
    answers: JSON.parse(row.answers_json) as AnswerInput[],
    createdAt: row.created_at,
  };
}

export async function enqueueSubmission(submission: PendingSubmission): Promise<void> {
  const db = await getDb();
  await db.runAsync(
    `INSERT OR REPLACE INTO pending_submissions
      (id, lesson_id, lesson_title, idempotency_key, answers_json, created_at)
     VALUES (?, ?, ?, ?, ?, ?)`,
    [
      submission.id,
      submission.lessonId,
      submission.lessonTitle,
      submission.idempotencyKey,
      JSON.stringify(submission.answers),
      submission.createdAt,
    ],
  );
}

export async function listPendingSubmissions(): Promise<PendingSubmission[]> {
  const db = await getDb();
  const rows = await db.getAllAsync<PendingRow>('SELECT * FROM pending_submissions ORDER BY created_at ASC');
  return rows.map(fromRow);
}

export async function removePendingSubmission(id: string): Promise<void> {
  const db = await getDb();
  await db.runAsync('DELETE FROM pending_submissions WHERE id = ?', [id]);
}
