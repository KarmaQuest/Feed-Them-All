// src/pages/ShopPage.tsx — Page boutique : items payants à acheter.
//
// Affiche uniquement les items du catalogue avec price_cents > 0.
// Les items gratuits à débloquer sont sur /quests.
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import { getCatalogue, getInventory, purchaseItem, type ShopItem, type InventoryItem } from '../api/shop'
import './ShopPage.css'

export default function ShopPage() {
  const { user } = useAuthStore()
  const navigate = useNavigate()

  const [items, setItems] = useState<ShopItem[]>([])
  const [inventory, setInventory] = useState<InventoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [purchasing, setPurchasing] = useState<string | null>(null)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')

  useEffect(() => {
    async function load() {
      setLoading(true)
      try {
        const [cat, inv] = await Promise.all([
          getCatalogue(),
          user ? getInventory() : Promise.resolve([] as InventoryItem[]),
        ])
        setItems(cat.filter(i => i.is_active && i.price_cents > 0))
        setInventory(inv)
      } catch {
        setError('Erreur chargement')
      }
      setLoading(false)
    }
    load()
  }, [user])

  const ownedSlugs = new Set(inventory.map(i => i.item.slug))

  async function handlePurchase(itemId: string) {
    if (!user) { navigate('/user-login'); return }
    setPurchasing(itemId)
    setError('')
    setSuccess('')
    try {
      const res = await purchaseItem(itemId)
      if (res.client_secret) {
        setSuccess('Redirection vers Stripe…')
      }
    } catch {
      setError('Erreur paiement')
    }
    setPurchasing(null)
  }

  const spriteSrc = (slug: string) => `/api/sprites/shop/${slug}/south.png`

  return (
    <div className="shop-page">
      <div className="shop-page__card-wrap">
        <div className="shop-page__header">
          <button className="shop-page__back" onClick={() => navigate('/', { state: { fromMap: true } })}>← Retour</button>
          <h1 className="shop-page__title">Boutique</h1>
        </div>

        {error && <div className="shop-page__alert shop-page__alert--error">{error}</div>}
        {success && <div className="shop-page__alert shop-page__alert--success">{success}</div>}

        {loading ? (
          <p className="shop-page__loading">Chargement…</p>
        ) : items.length === 0 ? (
          <p className="shop-page__empty">Aucun article en vente pour le moment</p>
        ) : (
          <section>
            <p className="shop-page__section-desc">
              Achète des skins et tenues premium pour soutenir l'application.
            </p>
            <div className="shop-page__items">
              {items.map(item => {
                const owned = ownedSlugs.has(item.slug)
                return (
                  <div key={item.id} className={`shop-page__card ${owned ? 'owned' : ''}`}>
                    <div className="shop-page__card-preview">
                      <img
                        src={spriteSrc(item.slug)}
                        alt={item.name}
                        className="shop-page__card-img"
                        onError={e => { (e.target as HTMLImageElement).style.display = 'none' }}
                      />
                    </div>
                    <div className="shop-page__card-info">
                      <h3 className="shop-page__card-name">{item.name}</h3>
                      <span className="shop-page__card-category">
                        {item.category === 'skin' ? 'Skin' : item.category === 'outfit' ? 'Tenue' : 'Accessoire'}
                      </span>
                      <span className="shop-page__card-price">
                        {(item.price_cents / 100).toFixed(2)} {item.currency.toUpperCase()}
                      </span>
                    </div>
                    <div className="shop-page__card-action">
                      {owned ? (
                        <span className="shop-page__owned-badge">✓ Possédé</span>
                      ) : (
                        <button
                          className="btn btn--style-yellow btn--full"
                          disabled={purchasing === item.id}
                          onClick={() => handlePurchase(item.id)}
                        >
                          {purchasing === item.id ? '…' : 'Acheter'}
                        </button>
                      )}
                    </div>
                  </div>
                )
              })}
            </div>
          </section>
        )}
      </div>
    </div>
  )
}
