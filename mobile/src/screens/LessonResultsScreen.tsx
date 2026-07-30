import { ScrollView, StyleSheet, Text, View } from 'react-native';
import type { NativeStackScreenProps } from '@react-navigation/native-stack';

import { DuoButton } from '../components/DuoButton';
import { colors, radii } from '../theme/colors';
import type { RootStackParamList } from '../navigation/types';

type Props = NativeStackScreenProps<RootStackParamList, 'LessonResults'>;

export function LessonResultsScreen({ route, navigation }: Props) {
  const { params } = route;

  const goHome = () => navigation.replace('Main');

  if (params.status === 'queued') {
    return (
      <View style={styles.container}>
        <Text style={styles.emoji}>📡</Text>
        <Text style={styles.title}>Saved offline</Text>
        <Text style={styles.subtitle}>
          &quot;{params.lessonTitle}&quot; will sync automatically once you&apos;re back online — your XP and streak will update then.
        </Text>
        <DuoButton label="Back to lessons" onPress={goHome} style={styles.button} />
      </View>
    );
  }

  const { result, lessonTitle } = params;
  const correctCount = result.results?.filter((r) => r.correct).length ?? 0;
  const totalCount = result.results?.length ?? 0;

  return (
    <ScrollView contentContainerStyle={styles.container}>
      <Text style={styles.emoji}>{result.score >= 80 ? '🎉' : '💪'}</Text>
      <Text style={styles.title}>{result.already_submitted ? 'Already completed' : 'Lesson complete!'}</Text>
      <Text style={styles.subtitle}>{lessonTitle}</Text>

      <View style={styles.statRow}>
        <View style={styles.statCard}>
          <Text style={styles.statValue}>{result.score}%</Text>
          <Text style={styles.statLabel}>Score</Text>
        </View>
        <View style={[styles.statCard, { backgroundColor: '#FFF7D6' }]}>
          <Text style={[styles.statValue, { color: colors.goldDark }]}>+{result.xp_earned}</Text>
          <Text style={styles.statLabel}>XP earned</Text>
        </View>
        <View style={[styles.statCard, { backgroundColor: '#FFE9E9' }]}>
          <Text style={[styles.statValue, { color: colors.red }]}>{result.hearts_remaining}</Text>
          <Text style={styles.statLabel}>Hearts left</Text>
        </View>
      </View>

      <View style={styles.streakRow}>
        <Text style={styles.streakText}>🔥 {result.current_streak}-day streak</Text>
      </View>

      {totalCount > 0 && (
        <View style={styles.breakdown}>
          <Text style={styles.breakdownTitle}>
            {correctCount} / {totalCount} correct
          </Text>
          {result.results!.map((r) => (
            <View key={r.exercise_id} style={styles.breakdownRow}>
              <Text style={styles.breakdownIcon}>{r.correct ? '✅' : '❌'}</Text>
              {!r.correct && <Text style={styles.breakdownAnswer}>{r.correct_answer}</Text>}
            </View>
          ))}
        </View>
      )}

      <DuoButton label="Continue" onPress={goHome} style={styles.button} />
    </ScrollView>
  );
}

const styles = StyleSheet.create({
  container: {
    flexGrow: 1,
    backgroundColor: colors.bg,
    alignItems: 'center',
    padding: 24,
    paddingTop: 72,
  },
  emoji: { fontSize: 64, marginBottom: 8 },
  title: { fontSize: 24, fontWeight: '800', color: colors.text },
  subtitle: { fontSize: 15, color: colors.textMuted, marginTop: 4, marginBottom: 24, textAlign: 'center' },
  statRow: { flexDirection: 'row', gap: 12, width: '100%', marginBottom: 20 },
  statCard: {
    flex: 1,
    backgroundColor: colors.bgMuted,
    borderRadius: radii.lg,
    paddingVertical: 16,
    alignItems: 'center',
  },
  statValue: { fontSize: 20, fontWeight: '800', color: colors.text },
  statLabel: { fontSize: 12, color: colors.textMuted, marginTop: 4 },
  streakRow: { marginBottom: 24 },
  streakText: { fontSize: 16, fontWeight: '700', color: colors.text },
  breakdown: { width: '100%', marginBottom: 24 },
  breakdownTitle: { fontSize: 14, fontWeight: '700', color: colors.textMuted, marginBottom: 8 },
  breakdownRow: { flexDirection: 'row', alignItems: 'center', gap: 8, paddingVertical: 4 },
  breakdownIcon: { fontSize: 16 },
  breakdownAnswer: { fontSize: 14, color: colors.textMuted },
  button: { width: '100%', marginTop: 'auto' },
});
