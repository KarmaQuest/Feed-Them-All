import './AvatarSprite.css'

export interface AvatarConfig {
  gender?: 'male' | 'female' | 'other'
  skin?: string
  outfit?: string
  accessory?: string
}

interface Props {
  config?: AvatarConfig | Record<string, unknown> | null
  size?: 'sm' | 'md' | 'lg'
  className?: string
}

function getSpritePath(config: AvatarConfig | Record<string, unknown> | null | undefined): string {
  const c = config as Record<string, unknown> | null | undefined
  if (!c) return ''
  const outfit = c?.outfit as string | undefined
  if (outfit) return `/api/sprites/shop/${outfit}/south.png`
  const gender = c?.gender
  const dir = gender === 'female' ? 'female' : 'male'
  return `/api/sprites/default/characters/${dir}/south.png`
}

const SIZE_MAP: Record<string, number> = {
  sm: 32,
  md: 48,
  lg: 64,
}

export default function AvatarSprite({ config, size = 'md', className = '' }: Props) {
  const src = getSpritePath(config)
  const px = SIZE_MAP[size]

  return src ? (
    <img
      src={src}
      alt="Avatar"
      className={`avs avs--${size} ${className}`}
      width={px}
      height={px}
      style={{ imageRendering: 'pixelated' }}
    />
  ) : null
}
