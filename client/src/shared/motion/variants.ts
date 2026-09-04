import type { Transition, Variants } from 'motion/react'

export const calmTransition: Transition = {
  duration: 0.42,
  ease: [0.22, 1, 0.36, 1],
}

export const revealUp: Variants = {
  hidden: { opacity: 0, y: 14 },
  visible: { opacity: 1, y: 0 },
}
