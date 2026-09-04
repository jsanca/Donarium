import { useTranslation } from 'react-i18next'
import artwork from '../../../../../../knowledge/design/wireframes/login/login-page.png'

export function LoginArtwork() {
  const { t } = useTranslation()

  return (
    <aside className="relative hidden overflow-hidden bg-surface-subtle lg:block">
      <img
        className="absolute inset-y-0 left-0 h-full max-w-none w-[170%] object-fill object-left"
        src={artwork}
        alt={t('login.artworkAlt')}
      />
      <div
        className="absolute inset-0 bg-linear-to-t from-canvas/45 via-transparent to-surface/20"
        aria-hidden="true"
      />
    </aside>
  )
}
