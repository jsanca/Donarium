import { House, Sprout } from 'lucide-react'
import { useTranslation } from 'react-i18next'

export function BrandSignature() {
  const { t } = useTranslation()

  return (
    <div className="flex flex-col items-center gap-3 text-ink">
      <span className="relative grid size-14 place-items-center" aria-hidden="true">
        <House className="size-14 stroke-[1.35]" />
        <Sprout className="absolute size-5 translate-y-1 stroke-[1.5]" />
      </span>
      <p className="font-display text-xs tracking-[0.3em] text-accent">
        {t('login.brandName')}
      </p>
    </div>
  )
}
