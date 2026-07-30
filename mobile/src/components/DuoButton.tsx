import { ActivityIndicator, Pressable, StyleSheet, Text, type StyleProp, type ViewStyle } from 'react-native';

import { colors, radii } from '../theme/colors';

type Variant = 'primary' | 'secondary' | 'danger' | 'outline';

interface DuoButtonProps {
  label: string;
  onPress: () => void;
  variant?: Variant;
  disabled?: boolean;
  loading?: boolean;
  style?: StyleProp<ViewStyle>;
}

const VARIANT_COLORS: Record<Variant, { bg: string; border: string; text: string }> = {
  primary: { bg: colors.green, border: colors.greenDark, text: '#FFFFFF' },
  secondary: { bg: colors.blue, border: colors.blueDark, text: '#FFFFFF' },
  danger: { bg: colors.red, border: colors.redDark, text: '#FFFFFF' },
  outline: { bg: '#FFFFFF', border: colors.border, text: colors.textMuted },
};

// Duolingo's signature "pressable 3D" button: a flat top color sitting on a
// darker bottom border that reads as depth. Disabled state flattens to gray.
export function DuoButton({ label, onPress, variant = 'primary', disabled, loading, style }: DuoButtonProps) {
  const palette = disabled ? { bg: colors.gray, border: '#D0D0D0', text: colors.grayDark } : VARIANT_COLORS[variant];

  return (
    <Pressable
      onPress={onPress}
      disabled={disabled || loading}
      style={({ pressed }) => [
        styles.base,
        { backgroundColor: palette.bg, borderColor: palette.border },
        pressed && !disabled && styles.pressed,
        style,
      ]}
    >
      {loading ? (
        <ActivityIndicator color={palette.text} />
      ) : (
        <Text style={[styles.label, { color: palette.text }]}>{label}</Text>
      )}
    </Pressable>
  );
}

const styles = StyleSheet.create({
  base: {
    height: 52,
    borderRadius: radii.lg,
    borderBottomWidth: 4,
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 16,
  },
  pressed: {
    borderBottomWidth: 2,
    marginTop: 2,
  },
  label: {
    fontSize: 16,
    fontWeight: '700',
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
});
