// Central route definitions shared by the UI and lightweight route tests.
export const paths = Object.freeze({
  home: '/',
  desks: '/desks',
  rooms: '/rooms',
  parking: '/parking',
  equipment: '/equipment',
  reservations: '/my-reservations',
  login: '/login',
  register: '/register',
  profile: '/profile',
  search: '/search',
  analytics: '/analytics',
  bookingConfirm: '/reservations/confirm',
  bookingSuccess: '/reservations/success',
  checkIn: '/reservations/:id/check-in',
  adminResources: '/admin/resources',
  createResource: '/admin/resources/new',
  editResource: '/admin/resources/:id/edit',
})

export const sections = Object.freeze([
  { path: paths.desks, label: 'Delovna mesta', description: 'Poišči prosto mizo.', icon: '▤' },
  { path: paths.rooms, label: 'Sejne sobe', description: 'Načrtuj naslednji sestanek.', icon: '▦' },
  { path: paths.parking, label: 'Parkirišča', description: 'Preveri parkirna mesta.', icon: '▧' },
  { path: paths.equipment, label: 'Oprema', description: 'Rezerviraj skupno opremo.', icon: '◈' },
])

export const navigation = Object.freeze([
  { path: paths.home, label: 'Domov', end: true },
  ...sections.map(({ path, label }) => ({ path, label, end: false })),
  { path: paths.reservations, label: 'Moje rezervacije', end: false },
  { path: paths.search, label: 'Iskanje virov', end: false },
  { path: paths.analytics, label: 'Analitika', end: false },
  { path: paths.login, label: 'Prijava', end: false },
  { path: paths.register, label: 'Registracija', end: false },
  { path: paths.profile, label: 'Moj profil', end: false },
  { path: paths.createResource, label: 'Dodaj vir', end: false },
  { path: paths.adminResources, label: 'Upravljanje virov', end: false },
])

