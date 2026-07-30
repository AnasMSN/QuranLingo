import { StyleSheet, Text, View } from 'react-native';

import { colors, radii } from '../theme/colors';

interface StatBadgeProps {
  icon: string;
  value: number | string;
  color?: string;
}

export function StatBadge({ icon, value, color = colors.text }: StatBadgeProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.icon}>{icon}</Text>
      <Text style={[styles.value, { color }]}>{value}</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.bgMuted,
    borderRadius: radii.pill,
    paddingHorizontal: 12,
    paddingVertical: 6,
    gap: 6,
  },
  icon: { fontSize: 16 },
  value: { fontSize: 15, fontWeight: '800' },
});
