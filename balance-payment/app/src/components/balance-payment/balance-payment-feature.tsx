import { useWallet } from '@solana/wallet-adapter-react'
import { WalletButton } from '../solana/solana-provider'
import { useBalancePaymentProgram } from './balance-payment-data-access'
import { useEffect, useState, useRef } from 'react'
import { BN } from '@coral-xyz/anchor'
import keccak from 'keccak'
import { useTransactionToast } from '../ui/ui-layout'
import toast from 'react-hot-toast'
import { finalizeEvent, generateSecretKey } from 'nostr-tools/pure'
import { Relay } from 'nostr-tools/relay'
import bs58 from 'bs58'
import ReactMarkdown from 'react-markdown'
import {
  addMessage,
  Conversation,
  createConversation,
  getAuthData,
  getConversations,
  getMessages,
  getUser,
  login,
  Message,
  User,
} from '@/services'

const SIGN_MESSAGE_PREFIX = 'DePHY vending machine/Example:\n'
const RELAY_ENDPOINT = import.meta.env.VITE_RELAY_ENDPOINT || 'ws://127.0.0.1:8000'

const MACHINE_PUBKEY = 'd041ea9854f2117b82452457c4e6d6593a96524027cd4032d2f40046deb78d93'

// define charge status
type ChargeStatus = 'idle' | 'requested' | 'working' | 'available' | 'error'

export default function BalancePaymentFeature() {
  const transactionToast = useTransactionToast()
  const { publicKey, wallet, signMessage } = useWallet()
  const { program, getGlobalPubkey, getUserAccountPubkey, generate64ByteUUIDPayload } = useBalancePaymentProgram()

  const [selectedTab, setSelectedTab] = useState<'payment' | 'chat'>('payment')
  const [serialNumberBytes, setSerialNumberBytes] = useState<Uint8Array | null>(null)
  const [globalAccount, setGlobalAccount] = useState<any>(null)
  const [userAccount, setUserAccount] = useState<any>(null)
  const [vaultBalance, setVaultBalance] = useState<number | null>(null)
  const [depositAmount, setDepositAmount] = useState<string>('')
  const [withdrawAmount, setWithdrawAmount] = useState<string>('')
  const [relay, setRelay] = useState<Relay>()
  const [sk, setSk] = useState<Uint8Array | null>(null)
  const [chargeStatus, setChargeStatus] = useState<ChargeStatus>('idle')
  const [events, setEvents] = useState<any[]>([])
  const [expandedEventIndex, setExpandedEventIndex] = useState<number | null>(null)
  const [isChargeDisabled, setIsChargeDisabled] = useState(false)
  const isTabDisabled = chargeStatus !== 'idle' && chargeStatus !== 'available'

  const [userInfo, setUserInfo] = useState<User | null>(null)
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [conversations, setConversations] = useState<Conversation[]>([])
  const [messages, setMessages] = useState<Partial<Message>[]>([])
  const [input, setInput] = useState('')
  const [isAskLoading, setIsAskLoading] = useState(false)
  const [aiModel, setAIModel] = useState<'deepseek/deepseek-v3/community' | 'deepseek/deepseek-r1/community'>(
    'deepseek/deepseek-v3/community',
  )
  const [isLogined, setIsLogined] = useState<boolean | null>(null)

  const subscriptionRef1 = useRef<any>(null)
  const messagesContainerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (messagesContainerRef.current) {
      messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight
    }
  }, [messages])

  useEffect(() => {
    const { uuidBytes } = generate64ByteUUIDPayload()
    setSerialNumberBytes(uuidBytes)
  }, [selectedTab])

  useEffect(() => {
    if (!globalAccount) {
      fetchGlobalAccount()
    }
  }, [program])

  useEffect(() => {
    if (!userAccount) {
      fetchUserAccount()
    }
  }, [program, publicKey])

  useEffect(() => {
    if (publicKey) {
      const authData = getAuthData()
      if (authData === null) {
        setIsLogined(false)
      } else {
        setIsLogined(true)
      }
    }
  }, [publicKey])

  useEffect(() => {
    ;(async () => {
      if (selectedTab === 'chat' && publicKey && isLogined) {
        if (conversationId) {
          const messagesWrapper = await getMessages(conversationId)
          if (messagesWrapper.error) {
            toast.error(`get messages failed, ${messagesWrapper.error}`)
            return
          }
          if (!messagesWrapper.data) {
            toast.error('get messages result undefined')
            return
          }
          setMessages(messagesWrapper.data)
        } else {
          const convosWrapper = await getConversations()
          if (convosWrapper.error) {
            toast.error(`get conversations failed, ${convosWrapper.error}`)
            return
          }
          if (!convosWrapper.data) {
            toast.error('get conversations result undefined')
            return
          }
          setConversations(convosWrapper.data)
          if (convosWrapper.data.length == 0) {
            const createWrapper = await createConversation()
            if (createWrapper.error) {
              toast.error(`create conversations failed, ${createWrapper.error}`)
              return
            }
            if (!createWrapper.data) {
              toast.error('create conversations result undefined')
              return
            }
            setConversationId(createWrapper.data.id)
          } else {
            setConversationId(convosWrapper.data[0].id)
          }
        }
      }
    })()
  }, [selectedTab, publicKey, isLogined, conversationId])

  useEffect(() => {
    if (publicKey && selectedTab === "payment") {
      const intervalId = setInterval(fetchUserAccount, 3000)

      return () => clearInterval(intervalId)
    }
  }, [publicKey, selectedTab])

  useEffect(() => {
    if (publicKey && isLogined && selectedTab === "payment") {
      const intervalId = setInterval(fetchUser, 3000)

      return () => clearInterval(intervalId)
    }
  }, [publicKey, isLogined, selectedTab])

  useEffect(() => {
    ;(async () => {
      const sk = generateSecretKey()
      setSk(sk)

      try {
        const relay = await Relay.connect(RELAY_ENDPOINT)
        setRelay(relay)
        toast.success(`connected to ${relay.url}`)
      } catch (error) {
        toast.error(`fail to connect relay, ${error}`)
      }
    })()
  }, [])

  const solToLamports = (sol: string): BN => {
    const solNumber = parseFloat(sol)
    if (isNaN(solNumber) || solNumber < 0) {
      throw new Error('Invalid SOL amount')
    }
    return new BN(solNumber * 10 ** 9)
  }

  const handleRegister = async () => {
    if (!publicKey || !program) {
      console.error('Wallet not connected or program not loaded')
      return
    }

    try {
      const transactionSignature = await program.methods
        .register()
        .accountsPartial({
          user: publicKey,
        })
        .rpc()

      console.log('Register transaction signature:', transactionSignature)
      transactionToast(transactionSignature)

      const userAccountPubkey = getUserAccountPubkey(publicKey)
      const user = await program.account.userAccount.fetch(userAccountPubkey)
      setUserAccount(user)
      const userVaultBalance = await program.provider.connection.getBalance(user.vault)
      setVaultBalance(userVaultBalance)
    } catch (error) {
      toast.error(`Error registering user account: ${error}`)
    }
  }

  const handleDeposit = async () => {
    if (!publicKey || !program || !depositAmount) {
      console.error('Wallet not connected or program not loaded or amount not set')
      return
    }

    try {
      const amount = solToLamports(depositAmount)
      const transactionSignature = await program.methods
        .deposit(amount)
        .accountsPartial({
          user: publicKey,
        })
        .rpc()

      console.log('Deposit transaction signature:', transactionSignature)
      transactionToast(transactionSignature)

      const userAccountPubkey = getUserAccountPubkey(publicKey)
      const user = await program.account.userAccount.fetch(userAccountPubkey)
      const userVaultBalance = await program.provider.connection.getBalance(user.vault)
      setVaultBalance(userVaultBalance)
    } catch (error) {
      toast.error(`Error depositing: ${error}`)
    }
  }

  const handleWithdraw = async () => {
    if (!publicKey || !program || !withdrawAmount) {
      console.error('Wallet not connected or program not loaded or amount not set')
      return
    }

    try {
      const amount = solToLamports(withdrawAmount)
      const transactionSignature = await program.methods
        .withdraw(amount)
        .accountsPartial({
          user: publicKey,
        })
        .rpc()

      console.log('Withdraw transaction signature:', transactionSignature)
      transactionToast(transactionSignature)

      const userAccountPubkey = getUserAccountPubkey(publicKey)
      const user = await program.account.userAccount.fetch(userAccountPubkey)
      const userVaultBalance = await program.provider.connection.getBalance(user.vault)
      setVaultBalance(userVaultBalance)
    } catch (error) {
      toast.error(`Error withdrawing: ${error}`)
    }
  }

  const handleSelectTab = (tab: 'payment' | 'chat') => {
    if (isTabDisabled) {
      return
    }
    setSelectedTab(tab)
    handleReset()
  }

  const fetchGlobalAccount = async () => {
    if (!program) return

    const globalAccountPubkey = getGlobalPubkey()
    const global = await program.account.globalAccount.fetch(globalAccountPubkey)
    setGlobalAccount(global)
  }

  const fetchUserAccount = async () => {
    if (!publicKey || !program) return

    const userAccountPubkey = getUserAccountPubkey(publicKey)
    const user = await program.account.userAccount.fetch(userAccountPubkey)
    const userVaultBalance = await program.provider.connection.getBalance(user.vault)
    setVaultBalance(userVaultBalance)
    setUserAccount(user)
  }

  const fetchUser = async () => {
    if (!publicKey || isLogined !== true) {
      return
    }
    const userInfoWrapper = await getUser()
    if (userInfoWrapper.error) {
      toast.error(`get user failed, ${userInfoWrapper.error}`)
      return
    }
    if (!userInfoWrapper.data) {
      toast.error('get user result undefined')
      return
    }
    setUserInfo(userInfoWrapper.data)
  }

  const handleCharge = async () => {
    if (!wallet || !publicKey || !signMessage || !serialNumberBytes) {
      console.error('Wallet not connected or serial number not generated')
      return
    }

    setIsChargeDisabled(true)

    const userAccountPubkey = getUserAccountPubkey(publicKey)
    const user = await program.account.userAccount.fetch(userAccountPubkey)

    const nonce = user.nonce
    const payload = serialNumberBytes
    const deadline = new BN(Date.now() / 1000 + 60 * 30) // 30 minutes later

    const message = Buffer.concat([payload, nonce.toArrayLike(Buffer, 'le', 8), deadline.toArrayLike(Buffer, 'le', 8)])
    const messageHash = keccak('keccak256').update(message).digest()
    const hashedMessageBase58 = bs58.encode(messageHash)
    const digest = new TextEncoder().encode(`${SIGN_MESSAGE_PREFIX}${hashedMessageBase58}`)

    let recoverInfo
    try {
      const signature = await signMessage(digest)
      recoverInfo = {
        signature: Array.from(signature),
        payload: Array.from(payload),
        deadline: deadline.toNumber(),
      }
    } catch (error) {
      toast.error(`Error signing message: ${error}`)
      setIsChargeDisabled(false)
      return
    }

    try {
      await publishToRelay(nonce.toNumber(), recoverInfo, publicKey.toString())
    } catch (error) {
      toast.error(`Error publishing to relay: ${error}`)
      setIsChargeDisabled(false)
      return
    }

    try {
      await listenFromRelay()
    } catch (error) {
      toast.error(`Error listening from relay: ${error}`)
      setIsChargeDisabled(false)
    }
  }

  const handleLogin = async () => {
    if (!wallet || !publicKey || !signMessage) {
      console.error('Wallet not connected or serial number not generated')
      return
    }
    const message = 'DePhy request to sign in'
    const digest = new TextEncoder().encode(message)

    const signature = await signMessage(digest)

    const signatureBase64 = Buffer.from(signature).toString('base64')

    const res = await login(publicKey.toString(), message, signatureBase64)

    console.log('login res:', res)

    setIsLogined(true)
  }

  const handleNewConversation = async () => {
    if (!publicKey) {
      console.error('Wallet not connected')
      return
    }
    const convoWrapper = await createConversation()
    if (convoWrapper.error) {
      toast.error(`create conversation failed, ${convoWrapper.error}`)
      return
    }
    if (!convoWrapper.data) {
      toast.error('create conversation result undefined')
      return
    }
    const convos = [convoWrapper.data, ...conversations]
    setConversations(convos)
    setConversationId(convoWrapper.data.id)
  }

  const handleSelectConversation = (id: string) => {
    setConversationId(id)
  }

  const handleDeepThink = () => {
    if (aiModel === 'deepseek/deepseek-v3/community') {
      setAIModel('deepseek/deepseek-r1/community')
    } else {
      setAIModel('deepseek/deepseek-v3/community')
    }
  }

  const handleAsk = async () => {
    if (!input.trim()) return
    if (!publicKey) {
      console.error('Wallet not connected')
      return
    }
    if (!conversationId) {
      console.error('conversationId not defined')
      return
    }
    setIsAskLoading(true)
    setInput('')
    setMessages((prev) => [...prev, { role: 'user', content: input }])

    setMessages((prev) => [...prev, { role: 'assistant', content: '' }])
    await addMessage(conversationId, input, aiModel, async (newContent: string) => {
      setMessages((prev) => {
        const newMessages = [...prev]
        const lastIndex = newMessages.length - 1

        // 防御性检查
        if (lastIndex < 0) return prev

        const lastMessage = newMessages[lastIndex]

        // 必须创建新对象保证不可变性
        if (lastMessage.role === 'assistant') {
          newMessages[lastIndex] = {
            ...lastMessage, // 展开原有属性
            content: lastMessage.content + newContent, // 创建新字符串
          }
        }

        return newMessages
      })
    })
    setIsAskLoading(false)
  }

  const publishToRelay = async (nonce: number, recoverInfo: any, user: string) => {
    if (!sk) {
      toast.error('sk not initialized')
      return
    }
    if (!relay) {
      toast.error('relay not initialized')
      return
    }
    const sTag = 'dephy-dsproxy-controller'

    const payload = JSON.stringify({
      recover_info: JSON.stringify(recoverInfo),
      nonce,
      user,
    })

    const contentData = {
      Request: {
        to_status: 'Working',
        reason: 'UserRequest',
        initial_request: '0000000000000000000000000000000000000000000000000000000000000000',
        payload,
      },
    }

    const content = JSON.stringify(contentData)

    let eventTemplate = {
      kind: 1573,
      created_at: Math.floor(Date.now() / 1000),
      tags: [
        ['s', sTag],
        ['p', MACHINE_PUBKEY],
      ],
      content,
    }
    const signedEvent = finalizeEvent(eventTemplate, sk)
    await relay.publish(signedEvent)
  }

  const listenFromRelay = async () => {
    if (!sk) {
      toast.error('sk not initialized')
      return
    }
    if (!relay) {
      toast.error('relay not initialized')
      return
    }

    const sTag = 'dephy-dsproxy-controller'

    // clear old subscription
    if (subscriptionRef1.current) {
      subscriptionRef1.current.close()
    }

    // create new subscription
    subscriptionRef1.current = relay.subscribe(
      [
        {
          kinds: [1573],
          since: Math.floor(Date.now() / 1000),
          '#s': [sTag],
          '#p': [MACHINE_PUBKEY],
        },
      ],
      {
        onevent: async (event) => {
          console.log('event received:', event)
          setEvents((prevEvents) => [...prevEvents, event])
          const content = JSON.parse(event.content)
          try {
            if (content.Request) {
              setChargeStatus('requested')
            } else if (content.Status) {
              if (content.Status.status === 'Working') {
                setChargeStatus('working')
              } else if (content.Status.status === 'Available') {
                setChargeStatus('available')
                setIsChargeDisabled(false)
              }
            }
          } catch (error) {
            console.error('Error parsing event content:', error)
            setChargeStatus('error')
            setEvents((prevEvents) => [...prevEvents, { error: 'Failed to parse event content', rawEvent: event }])
          }
        },
        oneose() {
          console.log('eose received')
        },
        onclose(reason) {
          console.log('close received:', reason)
        },
      },
    )
  }

  // reset status
  const handleReset = () => {
    // clear old subscription
    if (subscriptionRef1.current) {
      subscriptionRef1.current.close()
      subscriptionRef1.current = null
    }

    // setRecoverInfo(null)
    setEvents([])
    setChargeStatus('idle')
    setIsChargeDisabled(false)
  }

  const ProgressBar = () => {
    let progress = 0
    let statusText = ''
    let barColor = 'bg-gray-300' // default gray

    switch (chargeStatus) {
      case 'requested':
        progress = 33
        statusText = 'Requested - Waiting for ds_proxy station...'
        barColor = 'bg-blue-500'
        break
      case 'working':
        progress = 66
        statusText = 'Working - Pay in progress...'
        barColor = 'bg-blue-500'
        break
      case 'available':
        progress = 100
        statusText = 'Available - Pay completed!'
        barColor = 'bg-green-500'
        break
      case 'error':
        progress = 100
        statusText = 'Error - Something went wrong!'
        barColor = 'bg-red-500'
        break
      default:
        progress = 0
        statusText = 'Idle - Ready to pay'
        barColor = 'bg-gray-300'
    }

    return (
      <div className="w-full mb-8">
        <div className="h-2 w-full bg-gray-200 rounded-full overflow-hidden">
          <div className={`h-full ${barColor} transition-all duration-500`} style={{ width: `${progress}%` }}></div>
        </div>
        <p className="mt-2 text-sm text-gray-600">{statusText}</p>
      </div>
    )
  }

  const EventJsonViewer = ({ event, index }: { event: any; index: number }) => {
    const isExpanded = expandedEventIndex === index

    const toggleExpand = () => {
      if (isExpanded) {
        setExpandedEventIndex(null)
      } else {
        setExpandedEventIndex(index)
      }
    }

    const getPurpose = () => {
      const eventType = 'Deepseek Proxy'

      try {
        const content = JSON.parse(event.content)
        if (content.Request) return `${eventType} Request`
        if (content.Status) return `${eventType} Status: ${content.Status.status}`
      } catch {
        return 'Invalid Event'
      }
      return 'Unknown Event'
    }

    const formatTime = (timestamp: number) => {
      return new Date(timestamp * 1000).toLocaleString()
    }

    return (
      <div className="mt-4 p-4 bg-base-100 rounded-lg shadow-md">
        <div className="font-bold cursor-pointer flex justify-between items-center" onClick={toggleExpand}>
          <span>
            <span className="mr-2">{isExpanded ? '▲' : '▼'}</span>
            Event {index + 1} - {getPurpose()}
          </span>
          <div className="flex items-center">
            <span className="text-sm text-gray-500 mr-2">{formatTime(event.created_at)}</span>
          </div>
        </div>
        {isExpanded && (
          <pre className="mt-2 break-all whitespace-pre-wrap text-xs">{JSON.stringify(event, null, 2)}</pre>
        )}
      </div>
    )
  }

  return publicKey ? (
    <div className="max-w-4xl mx-auto p-4">
      <div className="flex justify-between items-center mb-8">
        {/* Tab */}
        <div className="inline-flex p-1 bg-gray-100 rounded-full">
          <button
            className={`px-6 py-2 rounded-full text-sm font-medium transition-all duration-300 ${
              selectedTab === 'payment' ? 'bg-white text-blue-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            } ${isTabDisabled ? 'opacity-50 cursor-not-allowed' : ''}`}
            onClick={() => handleSelectTab('payment')}
            disabled={isTabDisabled}
          >
            Payment
            {selectedTab === 'payment' && isTabDisabled && <span className="ml-2 animate-pulse">(Processing...)</span>}
          </button>
          <button
            className={`px-6 py-2 rounded-full text-sm font-medium transition-all duration-300 ${
              selectedTab === 'chat' ? 'bg-white text-pink-600 shadow-sm' : 'text-gray-500 hover:text-gray-700'
            } ${isTabDisabled ? 'opacity-50 cursor-not-allowed' : ''}`}
            onClick={() => handleSelectTab('chat')}
            disabled={isTabDisabled}
          >
            Chat
            {selectedTab === 'chat' && isTabDisabled && <span className="ml-2 animate-pulse">(Processing...)</span>}
          </button>
        </div>

        {/* Login Button */}
        <button
          className="btn px-8 py-2 rounded-full text-sm font-medium transition-all duration-300 bg-blue-600 text-white hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2"
          onClick={handleLogin} // 假设你有一个 handleLogin 函数
          disabled={isLogined === true}
        >
          Login
        </button>
      </div>

      {selectedTab === 'payment' ? (
        <>
          <div className="flex flex-wrap gap-4 mb-8">
            {/* userAccount */}
            <div className="flex-1 p-4 bg-base-200 rounded-lg shadow-md">
              <h2 className="text-xl font-bold mb-2">User Account</h2>
              {userAccount ? (
                <div className="space-y-2">
                  <p>
                    <span className="font-semibold text-sm">Nonce:</span>{' '}
                    <span className="text-sm">{userAccount.nonce.toString()}</span>
                  </p>
                  <p>
                    <span className="font-semibold text-sm">Locked Amount:</span>{' '}
                    <span className="text-sm">{userAccount.lockedAmount.toNumber() / 10 ** 9} SOL</span>
                  </p>
                  <p>
                    <span className="font-semibold text-sm">Vault:</span>{' '}
                    <span className="text-xs">{userAccount.vault.toString()}</span>
                  </p>
                  <p>
                    <span className="font-semibold text-sm">Vault Balance:</span>{' '}
                    <span className="text-sm">{vaultBalance ? `${vaultBalance / 10 ** 9} SOL` : 'Loading...'}</span>
                  </p>
                  <p>
                    <span className="font-semibold text-sm">Tokens:</span>{' '}
                    <span className="text-sm">{userInfo ? `${userInfo.tokens}` : 'Loading...'}</span>
                  </p>
                  <p>
                    <span className="font-semibold text-sm">Tokens Consumed:</span>{' '}
                    <span className="text-sm">{userInfo ? `${userInfo.tokens_consumed}` : 'Loading...'}</span>
                  </p>
                </div>
              ) : (
                <div className="flex flex-col items-center">
                  <p>No user account data found.</p>
                  <button className="btn btn-primary mt-4" onClick={handleRegister} disabled={!publicKey}>
                    Register
                  </button>
                </div>
              )}
            </div>

            {/* deposit */}
            <div className="flex-1 p-4 bg-base-200 rounded-lg shadow-md">
              <h2 className="text-xl font-bold mb-4">Deposit</h2>
              <div className="space-y-4">
                <input
                  type="text"
                  placeholder="Amount (SOL)"
                  value={depositAmount}
                  onChange={(e) => setDepositAmount(e.target.value)}
                  className="input input-bordered w-full placeholder:text-sm"
                />
                <button className="btn btn-primary w-full" onClick={handleDeposit} disabled={!depositAmount}>
                  Deposit
                </button>
              </div>
            </div>

            {/* withdraw */}
            <div className="flex-1 p-4 bg-base-200 rounded-lg shadow-md">
              <h2 className="text-xl font-bold mb-4">Withdraw</h2>
              <div className="space-y-4">
                <input
                  type="text"
                  placeholder="Amount (SOL)"
                  value={withdrawAmount}
                  onChange={(e) => setWithdrawAmount(e.target.value)}
                  className="input input-bordered w-full placeholder:text-sm"
                />
                <button className="btn btn-primary w-full" onClick={handleWithdraw} disabled={!withdrawAmount}>
                  Withdraw
                </button>
              </div>
            </div>
          </div>

          <div className="mb-8 p-4 bg-base-200 rounded-lg shadow-md">
            <h2 className="text-xl font-bold mb-4">{'Pay'}</h2>
            <ProgressBar />

            {events.map((event, index) => (
              <EventJsonViewer key={index} event={event} index={index} />
            ))}

            {chargeStatus === 'available' && (
              <button className="btn btn-secondary w-full mt-4" onClick={handleReset}>
                Reset
              </button>
            )}

            <button
              className={`btn btn-primary w-full mt-4 border-none bg-blue-500 hover:bg-blue-600 text-white`}
              onClick={handleCharge}
              disabled={!wallet || !serialNumberBytes || isChargeDisabled || chargeStatus !== 'idle'}
            >
              Pay
            </button>
          </div>
        </>
      ) : (
        <div className="w-full h-[70vh] flex">
          {/* 左侧对话列表 */}
          <div className="w-44 border-r p-4 overflow-auto flex flex-col">
            <div className="flex justify-between items-center mb-4">
              <button
                className="p-1 bg-pink-500 text-white rounded hover:bg-pink-600"
                onClick={handleNewConversation}
                aria-label="New Conversation"
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  className="h-5 w-5"
                  fill="none"
                  viewBox="0 0 24 24"
                  stroke="currentColor"
                  strokeWidth="2"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
                  />
                </svg>
              </button>
            </div>
            <div className="flex-1 space-y-2">
              {conversations.map((conversation) => (
                <div
                  key={conversation.id}
                  className={`p-1 rounded-lg cursor-pointer ${conversationId === conversation.id ? 'bg-gray-100' : ''}`}
                  onClick={() => handleSelectConversation(conversation.id)}
                >
                  {conversation.id}
                </div>
              ))}
            </div>
          </div>

          {/* 右侧聊天区域 */}
          <div className="flex-1 min-w-[700px] p-4 flex flex-col">
            {/* 消息列表 */}
            <div className="flex-1 overflow-auto space-y-4" ref={messagesContainerRef}>
              {messages.map((msg, index) => {
                const parts = msg.content!.split(/<think>(.*?)<\/think>/gs)
                return (
                  <div key={index} className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
                    <div
                      className={`p-3 rounded-lg shadow ${msg.role === 'user' ? 'bg-blue-100' : 'bg-gray-100 max-w-4xl w-full'}`}
                    >
                      {parts.map((part, i) =>
                        i % 2 === 0 ? (
                          <ReactMarkdown key={i}>{part}</ReactMarkdown>
                        ) : (
                          <span key={i} className="text-xs text-gray-500 block leading-tight mb-2">
                            {'>>'}
                            {part}
                          </span>
                        ),
                      )}
                    </div>
                  </div>
                )
              })}
            </div>

            {/* 输入框和按钮 */}
            <div className="mt-4 border-t pt-2">
              {/* 第一行：输入框 */}
              <div className="mb-2 w-full">
                <input
                  type="text"
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyDown={(e) => e.key === 'Enter' && handleAsk()}
                  className="w-full p-2 border rounded-lg"
                  placeholder="Ask DeepSeek"
                />
              </div>

              {/* 第二行：操作按钮组 */}
              <div className="flex justify-end items-center gap-4">
                {/* 模型切换标签 */}
                <div className="flex items-center">
                  <button
                    onClick={handleDeepThink}
                    className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors 
                      ${aiModel === 'deepseek/deepseek-r1/community' ? 'bg-pink-500' : 'bg-gray-300'}`}
                  >
                    <span
                      className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform
                      ${aiModel === 'deepseek/deepseek-r1/community' ? 'translate-x-6' : 'translate-x-1'}`}
                    />
                  </button>
                  <span className="ml-2 text-sm text-gray-600">DeepThink</span>
                </div>

                {/* Ask按钮 */}
                <button
                  onClick={handleAsk}
                  className="px-4 py-2 bg-pink-500 text-white rounded-lg flex items-center justify-center gap-2"
                  disabled={isAskLoading}
                >
                  {isAskLoading ? (
                    <svg
                      className="animate-spin h-5 w-5 text-white"
                      xmlns="http://www.w3.org/2000/svg"
                      fill="none"
                      viewBox="0 0 24 24"
                    >
                      <circle
                        className="opacity-25"
                        cx="12"
                        cy="12"
                        r="10"
                        stroke="currentColor"
                        strokeWidth="4"
                      ></circle>
                      <path
                        className="opacity-75"
                        fill="currentColor"
                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
                      ></path>
                    </svg>
                  ) : (
                    'Ask'
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  ) : (
    <div className="max-w-4xl mx-auto">
      <div className="hero py-[64px]">
        <div className="hero-content text-center">
          <WalletButton />
        </div>
      </div>
    </div>
  )
}
