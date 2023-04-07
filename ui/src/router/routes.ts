import {
  NavigationGuardNext,
  RouteLocationNormalized,
  RouteRecordRaw,
} from 'vue-router'
import Cookies from 'js-cookie'

export const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: 'home',
    component: () => import('@/pages/home'),
    beforeEnter: [requiredSignup],
  },
  {
    path: '/tenant',
    name: 'tenant',
    component: () => import('@/pages/tenant'),
    beforeEnter: [requiredSignup],
    children: [
      { path: 'user', name: 'user', component: () => import('@/pages/user') },
      { path: 'role', name: 'role', component: () => import('@/pages/role') },
      {
        path: 'permission',
        name: 'permission',
        component: () => import('@/pages/permission'),
      },
    ],
  },
  {
    path: '/signup',
    name: 'signup',
    component: () => import('@/pages/signup'),
  },
  {
    path: '/login',
    name: 'login',
    component: () => import('@/pages/login'),
  },
  {
    path: '/logout',
    name: 'logout',
    component: () => import('@/pages/logout'),
  },
  {
    path: '/unauthorized',
    name: 'unauthorized',
    component: () => import('@/pages/errors').then((v) => v.Unauthorized),
  },
  {
    path: '/:pathMatch(.*)*',
    name: 'notFound',
    component: () => import('@/pages/errors').then((v) => v.NotFound),
  },
]

function requiredSignup(
  to: RouteLocationNormalized,
  form: RouteLocationNormalized,
  next: NavigationGuardNext
) {
  const token = Cookies.get('token')
  if (!token) {
    return next({ name: 'signup' })
  }
  return next()
}
