import { Pressable, StyleSheet, Text, View } from 'react-native';

import { colors, radii } from '../theme/colors';
import type { LessonStatus } from '../types/api';

interface LessonNodeProps {
  status: LessonStatus;
  offsetX: number;
  color: string;
  onPress: () => void;
}

const NODE_SIZE = 72;

export function LessonNode({ status, offsetX, color, onPress }: LessonNodeProps) {
  const locked = status === 'locked';
  const completed = status === 'completed';

  const bg = locked ? colors.locked : completed ? colors.gold : color;
  const border = locked ? '#D0D0D0' : completed ? colors.goldDark : `${color}CC`;

  return (
    <View style={[styles.row, { transform: [{ translateX: offsetX }] }]}>
      <Pressable
        onPress={onPress}
        disabled={locked}
        style={({ pressed }) => [
          styles.node,
          { backgroundColor: bg, borderColor: border },
          pressed && !locked && styles.pressed,
        ]}
      >
        <Text style={styles.icon}>{locked ? '🔒' : completed ? '⭐' : '▶'}</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  row: {
    alignItems: 'center',
    marginVertical: 10,
  },
  node: {
    width: NODE_SIZE,
    height: NODE_SIZE,
    borderRadius: radii.pill,
    borderBottomWidth: 6,
    alignItems: 'center',
    justifyContent: 'center',
  },
  pressed: {
    borderBottomWidth: 2,
    marginTop: 4,
  },
  icon: { fontSize: 26 },
});
