declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module 'vue' {
  export * from '@vue/runtime-dom'
}

declare module 'vue-router' {
  export * from 'vue-router/dist/vue-router'
}

declare module 'ant-design-vue' {
  const antd: any
  export default antd
  export * from 'ant-design-vue/es'
}

declare module '@ant-design/icons-vue' {
  export const PlusOutlined: any
  export const EditOutlined: any
  export const DeleteOutlined: any
  export const SearchOutlined: any
  export const ReloadOutlined: any
  export const DownloadOutlined: any
  export const UploadOutlined: any
  export const EyeOutlined: any
  export const UserOutlined: any
  export const SettingOutlined: any
  export const BellOutlined: any
  export const ExclamationCircleOutlined: any
  export const CheckCircleOutlined: any
  export const CloseCircleOutlined: any
  export const InfoCircleOutlined: any
  export const WarningOutlined: any
  export const FilterOutlined: any
  export const ExportOutlined: any
  export const ImportOutlined: any
  export const CopyOutlined: any
  export const PlayCircleOutlined: any
  export const PauseCircleOutlined: any
  export const StopOutlined: any
  export const SyncOutlined: any
  export const CloudUploadOutlined: any
  export const FileTextOutlined: any
  export const TeamOutlined: any
  export const SafetyOutlined: any
  export const DatabaseOutlined: any
  export const ApiOutlined: any
  export const MailOutlined: any
  export const PhoneOutlined: any
  export const GlobalOutlined: any
  export const LockOutlined: any
  export const UnlockOutlined: any
  export const KeyOutlined: any
  export const HeartOutlined: any
  export const ThunderboltOutlined: any
  export const DashboardOutlined: any
  export const BarChartOutlined: any
  export const LineChartOutlined: any
  export const PieChartOutlined: any
  export const CalendarOutlined: any
  export const ClockCircleOutlined: any
  export const EnvironmentOutlined: any
  export const TagOutlined: any
  export const TagsOutlined: any
  export const FolderOutlined: any
  export const FileOutlined: any
  export const LinkOutlined: any
  export const DisconnectOutlined: any
  export const WifiOutlined: any
  export const CloudOutlined: any
  export const ServerOutlined: any
  export const MonitorOutlined: any
  export const MobileOutlined: any
  export const TabletOutlined: any
  export const DesktopOutlined: any
  export const PrinterOutlined: any
  export const ScanOutlined: any
  export const CameraOutlined: any
  export const VideoCameraOutlined: any
  export const AudioOutlined: any
  export const CustomerServiceOutlined: any
  export const QuestionCircleOutlined: any
  export const BugOutlined: any
  export const ToolOutlined: any
  export const BuildOutlined: any
  export const RocketOutlined: any
  export const GiftOutlined: any
  export const TrophyOutlined: any
  export const CrownOutlined: any
  export const StarOutlined: any
  export const LikeOutlined: any
  export const DislikeOutlined: any
  export const MessageOutlined: any
  export const CommentOutlined: any
  export const ChatOutlined: any
  export const NotificationOutlined: any
  export const SoundOutlined: any
  export const RadarChartOutlined: any
  export const QrcodeOutlined: any
  export const BarcodeOutlined: any
  export const NumberOutlined: any
  export const FontSizeOutlined: any
  export const FontColorsOutlined: any
  export const HighlightOutlined: any
  export const BoldOutlined: any
  export const ItalicOutlined: any
  export const UnderlineOutlined: any
  export const StrikethroughOutlined: any
  export const RedoOutlined: any
  export const UndoOutlined: any
  export const ZoomInOutlined: any
  export const ZoomOutOutlined: any
  export const FullscreenOutlined: any
  export const FullscreenExitOutlined: any
  export const PictureOutlined: any
  export const SaveOutlined: any
  export const FolderOpenOutlined: any
  export const FolderAddOutlined: any
  export const FileAddOutlined: any
  export const FileDoneOutlined: any
  export const FileExcelOutlined: any
  export const FilePdfOutlined: any
  export const FileWordOutlined: any
  export const FilePptOutlined: any
  export const FileImageOutlined: any
  export const FileZipOutlined: any
  export const FileUnknownOutlined: any
  export const FileProtectOutlined: any
  export const FileSearchOutlined: any
  export const FileSyncOutlined: any
  export const FileMarkdownOutlined: any
  export const ApartmentOutlined: any
  export const AuditOutlined: any
  export const BankOutlined: any
  export const CarOutlined: any
  export const CarryOutOutlined: any
  export const CloudDownloadOutlined: any
  export const CloudServerOutlined: any
  export const CloudSyncOutlined: any
  export const ContactsOutlined: any
  export const ContainerOutlined: any
  export const ControlOutlined: any
  export const CreditCardOutlined: any
  export const DollarOutlined: any
  export const EuroOutlined: any
  export const FundOutlined: any
  export const GoldOutlined: any
  export const HomeOutlined: any
  export const HourglassOutlined: any
  export const IdcardOutlined: any
  export const InsuranceOutlined: any
  export const InteractionOutlined: any
  export const LayoutOutlined: any
  export const LaptopOutlined: any
  export const MedicineBoxOutlined: any
  export const PayCircleOutlined: any
  export const PercentageOutlined: any
  export const ProfileOutlined: any
  export const ProjectOutlined: any
  export const PropertySafetyOutlined: any
  export const PushpinOutlined: any
  export const ReconciliationOutlined: any
  export const RedEnvelopeOutlined: any
  export const RestOutlined: any
  export const SafetyCertificateOutlined: any
  export const ScheduleOutlined: any
  export const SecurityScanOutlined: any
  export const SelectOutlined: any
  export const ShopOutlined: any
  export const ShoppingOutlined: any
  export const ShoppingCartOutlined: any
  export const SolutionOutlined: any
  export const SwitcherOutlined: any
  export const TabletOutlined: any
  export const TrademarkOutlined: any
  export const TransactionOutlined: any
  export const TruckOutlined: any
  export const UsbOutlined: any
  export const WalletOutlined: any
  export const BookOutlined: any
  export const AlertOutlined: any
  export const AimOutlined: any
  export const AppstoreOutlined: any
  export const MenuOutlined: any
  export const MenuFoldOutlined: any
  export const MenuUnfoldOutlined: any
  export const OrderedListOutlined: any
  export const UnorderedListOutlined: any
  export const BarsOutlined: any
  export const DotChartOutlined: any
  export const AreaChartOutlined: any
  export const StockOutlined: any
  export const BoxPlotOutlined: any
  export const SlackOutlined: any
  export const AlipayCircleOutlined: any
  export const TaobaoCircleOutlined: any
  export const WeiboCircleOutlined: any
  export const TwitterOutlined: any
  export const WechatOutlined: any
  export const YoutubeOutlined: any
  export const AlipayOutlined: any
  export const TaobaoOutlined: any
  export const WeiboSquareOutlined: any
  export const WeiboOutlined: any
  export const TwitterSquareOutlined: any
  export const FacebookOutlined: any
  export const SkypeOutlined: any
  export const CodeSandboxOutlined: any
  export const ChromeOutlined: any
  export const CodepenOutlined: any
  export const AliwangwangOutlined: any
  export const AppleOutlined: any
  export const AndroidOutlined: any
  export const WindowsOutlined: any
  export const IeOutlined: any
  export const GoogleOutlined: any
  export const AmazonOutlined: any
  export const SlackSquareOutlined: any
  export const BehanceOutlined: any
  export const BehanceSquareOutlined: any
  export const DribbbleOutlined: any
  export const DribbbleSquareOutlined: any
  export const InstagramOutlined: any
  export const YuqueOutlined: any
  export const AlibabaOutlined: any
  export const YahooOutlined: any
  export const RedditOutlined: any
  export const SketchOutlined: any
  export const AccountBookOutlined: any
  export const AlertTwoTone: any
  export const ApiTwoTone: any
  export const AppstoreTwoTone: any
  export const AudioTwoTone: any
  export const BankTwoTone: any
  export const BellTwoTone: any
  export const BookTwoTone: any
  export const BugTwoTone: any
  export const BuildTwoTone: any
  export const BulbTwoTone: any
  export const CalculatorTwoTone: any
  export const CalendarTwoTone: any
  export const CameraTwoTone: any
  export const CarTwoTone: any
  export const CarryOutTwoTone: any
  export const CheckCircleTwoTone: any
  export const CheckSquareTwoTone: any
  export const ClockCircleTwoTone: any
  export const CloseCircleTwoTone: any
  export const CloseSquareTwoTone: any
  export const CloudTwoTone: any
  export const CodeTwoTone: any
  export const CompassTwoTone: any
  export const ContactsTwoTone: any
  export const ContainerTwoTone: any
  export const ControlTwoTone: any
  export const CopyTwoTone: any
  export const CopyrightTwoTone: any
  export const CreditCardTwoTone: any
  export const CrownTwoTone: any
  export const CustomerServiceTwoTone: any
  export const DashboardTwoTone: any
  export const DatabaseTwoTone: any
  export const DeleteTwoTone: any
  export const DiffTwoTone: any
  export const DislikeTwoTone: any
  export const DollarCircleTwoTone: any
  export const DollarTwoTone: any
  export const DownCircleTwoTone: any
  export const DownSquareTwoTone: any
  export const EditTwoTone: any
  export const EnvironmentTwoTone: any
  export const EuroCircleTwoTone: any
  export const EuroTwoTone: any
  export const ExclamationCircleTwoTone: any
  export const ExperimentTwoTone: any
  export const EyeTwoTone: any
  export const EyeInvisibleTwoTone: any
  export const FileAddTwoTone: any
  export const FileExcelTwoTone: any
  export const FileExclamationTwoTone: any
  export const FileImageTwoTone: any
  export const FileMarkdownTwoTone: any
  export const FilePdfTwoTone: any
  export const FilePptTwoTone: any
  export const FileTextTwoTone: any
  export const FileTwoTone: any
  export const FileUnknownTwoTone: any
  export const FileWordTwoTone: any
  export const FileZipTwoTone: any
  export const FilterTwoTone: any
  export const FireTwoTone: any
  export const FlagTwoTone: any
  export const FolderAddTwoTone: any
  export const FolderOpenTwoTone: any
  export const FolderTwoTone: any
  export const FrownTwoTone: any
  export const FundTwoTone: any
  export const FunnelPlotTwoTone: any
  export const GiftTwoTone: any
  export const GoldTwoTone: any
  export const HeartTwoTone: any
  export const HighlightTwoTone: any
  export const HomeTwoTone: any
  export const HourglassTwoTone: any
  export const IdcardTwoTone: any
  export const InfoCircleTwoTone: any
  export const InsuranceTwoTone: any
  export const InteractionTwoTone: any
  export const LayoutTwoTone: any
  export const LeftCircleTwoTone: any
  export const LeftSquareTwoTone: any
  export const LikeTwoTone: any
  export const LockTwoTone: any
  export const MailTwoTone: any
  export const MedicineBoxTwoTone: any
  export const MehTwoTone: any
  export const MessageTwoTone: any
  export const MinusCircleTwoTone: any
  export const MinusSquareTwoTone: any
  export const MobileTwoTone: any
  export const MoneyCollectTwoTone: any
  export const NotificationTwoTone: any
  export const PauseCircleTwoTone: any
  export const PhoneTwoTone: any
  export const PictureTwoTone: any
  export const PieChartTwoTone: any
  export const PlayCircleTwoTone: any
  export const PlaySquareTwoTone: any
  export const PlusCircleTwoTone: any
  export const PlusSquareTwoTone: any
  export const PoundCircleTwoTone: any
  export const PrinterTwoTone: any
  export const ProfileTwoTone: any
  export const ProjectTwoTone: any
  export const PropertySafetyTwoTone: any
  export const PushpinTwoTone: any
  export const QuestionCircleTwoTone: any
  export const ReconciliationTwoTone: any
  export const RedEnvelopeTwoTone: any
  export const RestTwoTone: any
  export const RightCircleTwoTone: any
  export const RightSquareTwoTone: any
  export const RocketTwoTone: any
  export const SafetyCertificateTwoTone: any
  export const SafeTwoTone: any
  export const SaveTwoTone: any
  export const ScheduleTwoTone: any
  export const SecurityScanTwoTone: any
  export const SettingTwoTone: any
  export const ShopTwoTone: any
  export const ShoppingTwoTone: any
  export const SkinTwoTone: any
  export const SmileTwoTone: any
  export const SoundTwoTone: any
  export const StarTwoTone: any
  export const StopTwoTone: any
  export const SwitcherTwoTone: any
  export const TabletTwoTone: any
  export const TagTwoTone: any
  export const TagsTwoTone: any
  export const ThunderboltTwoTone: any
  export const ToolTwoTone: any
  export const TrademarkCircleTwoTone: any
  export const TrophyTwoTone: any
  export const UnlockTwoTone: any
  export const UpCircleTwoTone: any
  export const UpSquareTwoTone: any
  export const UsbTwoTone: any
  export const VideoCameraTwoTone: any
  export const WalletTwoTone: any
  export const WarningTwoTone: any
}

declare module '@/utils/datetime' {
  export function formatDateTime(date: string | Date): string
  export function formatDate(date: string | Date): string
  export function formatTime(date: string | Date): string
  export function getRelativeTime(date: string | Date): string
  export function isToday(date: string | Date): boolean
  export function isYesterday(date: string | Date): boolean
  export function getDaysAgo(days: number): Date
  export function getHoursAgo(hours: number): Date
  export function getMinutesAgo(minutes: number): Date
}

declare module '@/services/*' {
  const service: any
  export default service
  export * from '@/services/*'
}

declare module '@/components/*' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '@/views/*' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

declare module '@/router' {
  import type { Router } from 'vue-router'
  const router: Router
  export default router
}

declare module '@/store' {
  const store: any
  export default store
}

// Global types
declare global {
  interface Window {
    __VUE_DEVTOOLS_GLOBAL_HOOK__?: any
  }
}

// NodeJS namespace for setTimeout, setInterval, etc.
declare namespace NodeJS {
  interface Timeout {}
  interface Timer {}
}

// Promise polyfill for older environments
interface PromiseConstructor {
  new <T>(executor: (resolve: (value?: T | PromiseLike<T>) => void, reject: (reason?: any) => void) => void): Promise<T>
  resolve<T>(value?: T | PromiseLike<T>): Promise<T>
  reject<T = never>(reason?: any): Promise<T>
  all<T>(values: readonly (T | PromiseLike<T>)[]): Promise<T[]>
  race<T>(values: readonly (T | PromiseLike<T>)[]): Promise<T>
}

declare var Promise: PromiseConstructor

// Object polyfills
interface ObjectConstructor {
  assign<T, U>(target: T, source: U): T & U
  assign<T, U, V>(target: T, source1: U, source2: V): T & U & V
  assign<T, U, V, W>(target: T, source1: U, source2: V, source3: W): T & U & V & W
  assign(target: object, ...sources: any[]): any
  keys(o: object): string[]
  values<T>(o: { [s: string]: T } | ArrayLike<T>): T[]
  entries<T>(o: { [s: string]: T } | ArrayLike<T>): [string, T][]
}

// Array polyfills
interface Array<T> {
  includes(searchElement: T, fromIndex?: number): boolean
  find<S extends T>(predicate: (this: void, value: T, index: number, obj: T[]) => value is S, thisArg?: any): S | undefined
  find(predicate: (value: T, index: number, obj: T[]) => unknown, thisArg?: any): T | undefined
  findIndex(predicate: (value: T, index: number, obj: T[]) => unknown, thisArg?: any): number
}

// String polyfills
interface String {
  includes(searchString: string, position?: number): boolean
  startsWith(searchString: string, position?: number): boolean
  endsWith(searchString: string, length?: number): boolean
  repeat(count: number): string
  padStart(targetLength: number, padString?: string): string
  padEnd(targetLength: number, padString?: string): string
}

export {}