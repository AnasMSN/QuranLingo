import { useQuery } from '@tanstack/react-query';

import { api } from '../api/client';
import { getCachedCourse, getCachedLesson, setCachedCourse, setCachedLesson } from '../db/courseCache';
import type { CourseTree, LessonDetail } from '../types/api';

const COURSE_CODE = 'quran-125';

export function useCourse() {
  return useQuery<CourseTree>({
    queryKey: ['course', COURSE_CODE],
    queryFn: async () => {
      try {
        const tree = await api.getCourse(COURSE_CODE);
        await setCachedCourse(COURSE_CODE, tree);
        return tree;
      } catch (err) {
        const cached = await getCachedCourse(COURSE_CODE);
        if (cached) return cached;
        throw err;
      }
    },
    staleTime: 60_000,
  });
}

export function useLessonDetail(lessonId: string) {
  return useQuery<LessonDetail>({
    queryKey: ['lesson', lessonId],
    queryFn: async () => {
      try {
        const detail = await api.getLesson(lessonId);
        await setCachedLesson(lessonId, detail);
        return detail;
      } catch (err) {
        const cached = await getCachedLesson(lessonId);
        if (cached) return cached;
        throw err;
      }
    },
    staleTime: 60_000,
  });
}
