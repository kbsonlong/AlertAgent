import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

// 扩展dayjs插件
dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

/**
 * 格式化日期
 * @param date 日期字符串或Date对象
 * @param format 格式化模板，默认为 'YYYY-MM-DD HH:mm:ss'
 * @returns 格式化后的日期字符串
 */
export const formatDate = (date?: string | Date, format = 'YYYY-MM-DD HH:mm:ss'): string => {
  if (!date) return '-'
  return dayjs(date).format(format)
}

/**
 * 格式化相对时间
 * @param date 日期字符串或Date对象
 * @returns 相对时间字符串，如"2小时前"
 */
export const formatRelativeTime = (date?: string | Date): string => {
  if (!date) return '-'
  return dayjs(date).fromNow()
}

/**
 * 格式化日期为简短格式
 * @param date 日期字符串或Date对象
 * @returns 简短格式的日期字符串
 */
export const formatDateShort = (date?: string | Date): string => {
  if (!date) return '-'
  const now = dayjs()
  const target = dayjs(date)
  
  if (now.isSame(target, 'day')) {
    return target.format('HH:mm')
  } else if (now.isSame(target, 'year')) {
    return target.format('MM-DD HH:mm')
  } else {
    return target.format('YYYY-MM-DD')
  }
}

/**
 * 判断日期是否为今天
 * @param date 日期字符串或Date对象
 * @returns 是否为今天
 */
export const isToday = (date?: string | Date): boolean => {
  if (!date) return false
  return dayjs().isSame(dayjs(date), 'day')
}

/**
 * 判断日期是否为昨天
 * @param date 日期字符串或Date对象
 * @returns 是否为昨天
 */
export const isYesterday = (date?: string | Date): boolean => {
  if (!date) return false
  return dayjs().subtract(1, 'day').isSame(dayjs(date), 'day')
}

/**
 * 获取时间差（毫秒）
 * @param startDate 开始日期
 * @param endDate 结束日期，默认为当前时间
 * @returns 时间差（毫秒）
 */
export const getTimeDiff = (startDate: string | Date, endDate?: string | Date): number => {
  const start = dayjs(startDate)
  const end = endDate ? dayjs(endDate) : dayjs()
  return end.diff(start)
}

/**
 * 格式化持续时间
 * @param duration 持续时间（毫秒）
 * @returns 格式化后的持续时间字符串
 */
export const formatDuration = (duration: number): string => {
  if (duration < 1000) {
    return `${duration}ms`
  } else if (duration < 60000) {
    return `${Math.round(duration / 1000)}s`
  } else if (duration < 3600000) {
    return `${Math.round(duration / 60000)}m`
  } else {
    return `${Math.round(duration / 3600000)}h`
  }
}