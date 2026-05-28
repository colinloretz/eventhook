import { Routes, Route } from 'react-router-dom'
import { Layout } from './components/Layout'
import { EventsPage } from './pages/EventsPage'
import { EventDetailPage } from './pages/EventDetailPage'
import { DeliveriesPage } from './pages/DeliveriesPage'
import { DeliveryDetailPage } from './pages/DeliveryDetailPage'
import { EndpointsPage } from './pages/EndpointsPage'

export default function App() {
  return (
    <Routes>
      <Route element={<Layout />}>
        <Route index element={<EventsPage />} />
        <Route path="events/:id" element={<EventDetailPage />} />
        <Route path="deliveries" element={<DeliveriesPage />} />
        <Route path="deliveries/:id" element={<DeliveryDetailPage />} />
        <Route path="endpoints" element={<EndpointsPage />} />
      </Route>
    </Routes>
  )
}
