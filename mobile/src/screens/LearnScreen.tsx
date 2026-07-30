import { useMemo, useState } from 'react';
import { ActivityIndicator, RefreshControl, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';
import type { CompositeScreenProps } from '@react-navigation/native';
import type { BottomTabScreenProps } from '@react-navigation/bottom-tabs';

import { LessonNode } from '../components/LessonNode';
import { StatBadge } from '../components/StatBadge';
import { useCourse } from '../hooks/useCourse';
import { useAuthStore } from '../store/authStore';
import { colors } from '../theme/colors';
import type { MainTabParamList, RootStackParamList } from '../navigation/types';

type Props = CompositeScreenProps<
  BottomTabScreenProps<MainTabParamList, 'Learn'>,
  NativeStackScreenProps<RootStackParamList>
>;

const SKILL_COLORS = [colors.green, colors.blue, colors.purple, colors.gold, colors.red];
const OFFSETS = [-50, 0, 50, 0];

export function LearnScreen({ navigation }: Props) {
  const { data: course, isLoading, error, refetch, isRefetching } = useCourse();
  const user = useAuthStore((s) => s.user);
  const [lockedHint, setLockedHint] = useState(false);

  // Precomputed once per course payload so each lesson node gets a stable
  // zigzag offset without mutating state during render. Must run before any
  // early return below so hooks stay unconditional across renders.
  const lessonOffsets = useMemo(() => {
    const offsets = new Map<string, number>();
    let i = 0;
    for (const skill of course?.skills ?? []) {
      for (const lesson of skill.lessons) {
        offsets.set(lesson.id, OFFSETS[i % OFFSETS.length]);
        i += 1;
      }
    }
    return offsets;
  }, [course]);

  if (isLoading) {
    return (
      <View style={styles.center}>
        <ActivityIndicator size="large" color={colors.green} />
      </View>
    );
  }

  if (error || !course) {
    return (
      <View style={styles.center}>
        <Text style={styles.errorText}>Couldn&apos;t load your course.</Text>
        <Text style={styles.errorSubtext}>Check your connection and pull down to retry.</Text>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <View style={styles.header}>
        <Text style={styles.title}>{course.title}</Text>
        <View style={styles.stats}>
          <StatBadge icon="❤️" value={user?.hearts ?? '-'} />
          <StatBadge icon="🔥" value={user?.current_streak ?? 0} color={colors.red} />
          <StatBadge icon="⭐" value={user?.total_xp ?? 0} color={colors.goldDark} />
        </View>
      </View>

      {lockedHint && (
        <View style={styles.hintBanner}>
          <Text style={styles.hintText}>Complete the lesson before this one first 🔒</Text>
        </View>
      )}

      <ScrollView
        contentContainerStyle={styles.scroll}
        refreshControl={<RefreshControl refreshing={isRefetching} onRefresh={refetch} />}
      >
        {course.skills.map((skill, skillIdx) => (
          <View key={skill.id} style={styles.skillSection}>
            <View style={styles.skillBanner}>
              <Text style={styles.skillIcon}>{skill.icon}</Text>
              <View style={styles.skillTextWrap}>
                <Text style={styles.skillTitle}>{skill.title}</Text>
                <Text style={styles.skillDescription}>{skill.description}</Text>
              </View>
            </View>

            {skill.lessons.map((lesson) => {
              return (
                <LessonNode
                  key={lesson.id}
                  status={lesson.status}
                  offsetX={lessonOffsets.get(lesson.id) ?? 0}
                  color={SKILL_COLORS[skillIdx % SKILL_COLORS.length]}
                  onPress={() => {
                    if (lesson.status === 'locked') {
                      setLockedHint(true);
                      setTimeout(() => setLockedHint(false), 2000);
                      return;
                    }
                    navigation.navigate('Lesson', { lessonId: lesson.id, lessonTitle: lesson.title });
                  }}
                />
              );
            })}
          </View>
        ))}
        <View style={{ height: 40 }} />
      </ScrollView>
    </View>
  );
}

const styles = StyleSheet.create({
  container: { flex: 1, backgroundColor: colors.bg },
  center: { flex: 1, alignItems: 'center', justifyContent: 'center', padding: 24, gap: 8 },
  errorText: { fontSize: 16, fontWeight: '700', color: colors.text },
  errorSubtext: { fontSize: 14, color: colors.textMuted },
  header: {
    paddingTop: 56,
    paddingHorizontal: 20,
    paddingBottom: 12,
    borderBottomWidth: 1,
    borderBottomColor: colors.border,
  },
  title: { fontSize: 22, fontWeight: '800', color: colors.text, marginBottom: 10 },
  stats: { flexDirection: 'row', gap: 10 },
  hintBanner: { backgroundColor: colors.gold, padding: 10, alignItems: 'center' },
  hintText: { fontWeight: '700', color: '#5C4400' },
  scroll: { paddingVertical: 24, alignItems: 'center' },
  skillSection: { width: '100%', alignItems: 'center', marginBottom: 16 },
  skillBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.bgMuted,
    borderRadius: 16,
    padding: 14,
    marginHorizontal: 24,
    marginBottom: 12,
    width: '85%',
    gap: 12,
  },
  skillIcon: { fontSize: 28 },
  skillTextWrap: { flex: 1 },
  skillTitle: { fontSize: 16, fontWeight: '800', color: colors.text },
  skillDescription: { fontSize: 12, color: colors.textMuted, marginTop: 2 },
});
