import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

// 配置dayjs
dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

/**
 * 格式化日期时间
 * @param date 日期字符串或Date对象
 * @param format 格式化模板，默认为 'YYYY-MM-DD HH:mm:ss'
 * @returns 格式化后的日期字符串
 */
export const formatDateTime = (date: string | Date, format = 'YYYY-MM-DD HH:mm:ss'): string => {
  if (!date) return '-'
  return dayjs(date).format(format)
}

/**
 * 格式化日期
 * @param date 日期字符串或Date对象
 * @returns 格式化后的日期字符串
 */
export const formatDate = (date: string | Date): string => {
  return formatDateTime(date, 'YYYY-MM-DD')
}

/**
 * 格式化时间
 * @param date 日期字符串或Date对象
 * @returns 格式化后的时间字符串
 */
export const formatTime = (date: string | Date): string => {
  return formatDateTime(date, 'HH:mm:ss')
}

/**
 * 获取相对时间
 * @param date 日期字符串或Date对象
 * @returns 相对时间字符串，如"2小时前"
 */
export const getRelativeTime = (date: string | Date): string => {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

/**
 * 判断是否为今天
 * @param date 日期字符串或Date对象
 * @returns 是否为今天
 */
export const isToday = (date: string | Date): boolean => {
  return dayjs(date).isSame(dayjs(), 'day')
}

/**
 * 判断是否为昨天
 * @param date 日期字符串或Date对象
 * @returns 是否为昨天
 */
export const isYesterday = (date: string | Date): boolean => {
  return dayjs(date).isSame(dayjs().subtract(1, 'day'), 'day')
}

/**
 * 获取友好的时间显示
 * @param date 日期字符串或Date对象
 * @returns 友好的时间字符串
 */
export const getFriendlyTime = (date: string | Date): string => {
  if (!date) return '-'
  
  const target = dayjs(date)
  const now = dayjs()
  
  if (target.isSame(now, 'day')) {
    return `今天 ${target.format('HH:mm')}`
  } else if (target.isSame(now.subtract(1, 'day'), 'day')) {
    return `昨天 ${target.format('HH:mm')}`
  } else if (target.isSame(now, 'year')) {
    return target.format('MM-DD HH:mm')
  } else {
    return target.format('YYYY-MM-DD HH:mm')
  }
}

/**
 * 计算时间差（秒）
 * @param startDate 开始时间
 * @param endDate 结束时间，默认为当前时间
 * @returns 时间差（秒）
 */
export const getTimeDiff = (startDate: string | Date, endDate: string | Date = new Date()): number => {
  return dayjs(endDate).diff(dayjs(startDate), 'second')
}

/**
 * 格式化持续时间
 * @param seconds 秒数
 * @returns 格式化后的持续时间字符串
 */
export const formatDuration = (seconds: number): string => {
  if (seconds < 60) {
    return `${seconds}秒`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    const remainingSeconds = seconds % 60
    return remainingSeconds > 0 ? `${minutes}分${remainingSeconds}秒` : `${minutes}分钟`
  } else if (seconds < 86400) {
    const hours = Math.floor(seconds / 3600)
    const remainingMinutes = Math.floor((seconds % 3600) / 60)
    return remainingMinutes > 0 ? `${hours}小时${remainingMinutes}分钟` : `${hours}小时`
  } else {
    const days = Math.floor(seconds / 86400)
    const remainingHours = Math.floor((seconds % 86400) / 3600)
    return remainingHours > 0 ? `${days}天${remainingHours}小时` : `${days}天`
  }
}

export default dayjs