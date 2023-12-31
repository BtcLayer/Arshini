<!-- @format -->
<script>
  import LuaCode from './LuaCode.svelte'

  let faucetcode = `function __init__ ()
  return {}
end

function fundfaucet ()
  -- don't need to do anything here, any money you send will just be added
  -- to the total contract funds.
end

function getmoney ()
  -- here we require the user to be authenticated.
  if not account.id then
    -- this terminates the call execution and
    -- invalidates anything that might have happened before.
    error('you must be authenticated')
  end

  -- we let the user specify how much he's going to take from the faucet
  local money_to_get = call.payload.money_to_get
  if not money_to_get then
    money_to_get = 10000 -- default to 10 sat
  end
  if money_to_get > 100000 then
    money_to_get = 100000 -- the amount is limited to 100 sat
  end

  -- this sends the money to the account balance of the caller
  contract.send(account.id, money_to_get)
end`
</script>

<main>
  <h1 id="writing-a-contract">Writing a contract</h1>
  <p>
    A contract consists of a <strong>state</strong>, some
    <strong>funds</strong> and multiple <strong>methods</strong>, which can be
    called by anyone and may affect the contract state, make GET requests to
    other places on the internet and manage the contract funds. What we call
    <em>methods</em> are just Lua functions.
  </p>
  <p>See the following example of a simple "community faucet" contract code:</p>
  <LuaCode>{faucetcode}</LuaCode>
  <p>
    The contract above allows generous rich people to put money in (by calling
    <code>fundfaucet</code> and including some satoshis in the call) and poor
    people to get the money out (by calling <code>getmoney</code> and specifying
    how much they want to withdraw in the call payload).
  </p>
  <p>
    However, it is very naïve and could easily be exploited. Although it limits
    withdrawals to 100 sat, someone could easily call the
    <code>getmoney</code> method multiple times and get all the money to itself.
    A less naïve approach would use the global <code>contract.state</code> to
    store a list of accounts who have already called <code>getmoney</code> and
    disallow them from calling it again. Another approach would be to use
    <code>os.time()</code> to keep track of the time of the withdrawals and
    force people to wait some time until they can call
    <code>getmoney</code> again. The possibilities are endless.
  </p>
  <p>
    Now that you've seen a contract, here are other things you must know about
    them:
  </p>
  <ul>
    <li>
      Each contract must have an <code>__init__</code> method. That is a special
      method that is called only when the contract is created, it must return a Lua
      table that will serve as the initial contract state.
    </li>
    <li>
      All other top level functions are methods callable from the external
      world, except methods with names beginning in an underscore:
      <code>_</code>.
    </li>
    <li>
      Some internal functions called from within a contract may fail and these
      may call the call execution to terminate, like
      <code>contract.send()</code>. Others, like <code>http</code> calls, can
      fail without causing the contract to terminate, instead they return an
      error, If you want the call to fail completely you must check for these
      errors and call the Lua function <code>error()</code> directly.
    </li>
    <li>
      All calls and payloads will be stored publicly in the contract history,
      except for errored calls.
    </li>
    <li>
      For your contracts to be easily integrated into the contract explorer and
      method caller interface in this website, make sure to follow these
      guidelines:
      <ul>
        <li>
          Write top-level methods with 0 indentation, both the
          <code>function</code> and the <code>end</code> keywords.
        </li>
        <li>
          When referring to payload fields, use the full object path, like
          <code>call.payload.fieldname</code> at least once, in other words,
          don't assign <code>call.payload</code> to another variable or you'll break
          our naïve regex-based parser.
        </li>
        <li>
          If you intend to use helper functions from inside the main methods,
          don't rely on <code>call</code> and <code>contract</code> globals, instead
          pass them as arguments.
        </li>
      </ul>
      These are just soft requirements and they may be dropped once we get a better
      Lua parser on our UI (but that will probably never happen).
    </li>
    <li>
      No one is able to change a contract's code after it has been activated,
      not even the contract creator (but contracts can be deleted if you made a
      mistake when creating them, provided they're new and don't have any
      funds). But of course this is a centralized system and the Etleneum team
      may delete contracts if they consider them wrong or harmful in any way.
    </li>
  </ul>
  <h1 id="calling-a-contract">Calling a contract</h1>
  <p>When you make a call, you send 4 things to the contract:</p>
  <ul>
    <li>
      A string <strong>method</strong> with the name of the contract method you're
      calling.
    </li>
    <li>A JSON <strong>payload</strong>.</li>
    <li>
      An integer representing the number of <strong>msatoshis</strong> to
      include in your call. Some methods may require you to include a certain
      number of msatoshis so they can be effective. The invoice you're required
      to pay to make any call includes this number of msatoshis plus a small
      antispam cost.
      <br />
      Regardless of what the contract code does with them, the msatoshis are always
      added to the contract funds.
    </li>
    <li>
      Optionally, a <code>?session=&lt;String&gt;</code> query string identifying
      the current authenticated user,
    </li>
  </ul>
  <h1 id="contract-api">Contract API</h1>
  <p>Contract code has access to the following globals:</p>
  <ul>
    <li>
      <code>call</code> table with fields:
      <ul>
        <li><code>id: String</code>, the call id, mostly useless;</li>
        <li>
          <code>payload: Any</code>, the payload submitted along with the call;
        </li>
        <li><code>msatoshi: Int</code>, the funds included in the call;</li>
      </ul>
    </li>
    <li>
      <code>contract</code> table with fields:
      <ul>
        <li><code>id: String</code>, the contract id, mostly useless;</li>
        <li>
          <code>state: Any</code>, the contract current state, should be mutated
          in-place;
        </li>
        <li>
          <code>get_funds: () => Int</code>, a function that returns the
          contract's current funds, in msatoshi;
        </li>
        <li>
          <code>send: (target: String, msatoshi: Int) => ()</code>, a function
          that sends from the contract funds to an user/contract;
        </li>
      </ul>
    </li>
    <li>
      <code>account</code> table with fields:
      <ul>
        <li>
          <code>id: String</code>, the account id of the caller,
          <code>nil</code> when the call is not authenticated;
        </li>
        <li>
          <code>get_balance: () => Int</code>, a function that returns the
          caller's full balance;
        </li>
        <li>
          <code>send: (target: String, msatoshi: Int) => ()</code>, a function
          that sends from the caller's balance to another user/contract;
        </li>
      </ul>
    </li>
    <li>
      <code>etleneum</code> table with fields:
      <ul>
        <li>
          <code>get_contract: (id: String) => (state: Any, funds: Int)</code>, a
          function that returns data about another contract.
        </li>
        <li>
          <code
            >call_external: (contract_id: String, method: String, payload: Any,
            msatoshi: Int) => ()</code
          >, a function that calls a method on another contract -- each external
          call costs one satoshi and it must be manually included in the current
          call;
        </li>
      </ul>
    </li>
    <li>
      <code>util</code> table with functions:
      <ul>
        <li>
          <code>print: (...Any) => ()</code>, shows a notification to the caller
          if he is listening to the server stream when making the call;
        </li>
        <li>
          <code>sha256: (any: String) => String</code>, takes any string and
          returns it's SHA256, hex-encoded.
        </li>
        <li><code>cuid: () => String</code>, generates a random id;</li>
        <li>
          <code
            >parse_bolt11: (pr: String) => (inv: &#123;payee: String, expiry:
            Int, routes: Array, currency: String, msatoshi: Int, created_at:
            Int, description: String, payment_hash: String,
            min_final_cltv_expiry: Int&#125;, error)</code
          >, parses a bolt11 invoice (and checks signatures, as this has not
          many more uses besides checking signatures);
        </li>
        <li>
          <code>check_address: (addr: String) => error</code>, checks if the
          given string is a valid Bitcoin address, returns nil if it is valid;
        </li>
      </ul>
    </li>
    <li>
      <code>http</code> table with functions:
      <ul>
        <li>
          <code
            >gettext: (url: String, headers: Map) => (body: String, error)</code
          >, calls an URL and returns the body response as text;
        </li>
        <li>
          <code>getjson: (url: String, headers: Map) => (body: Any, error)</code
          >, calls an URL and returns the body response JSON as a table;
        </li>
        <li>
          <code
            >postjson: (url: String, data: Any, headers: Map) => (body: Any,
            error)</code
          >, calls an URL and returns the body response JSON as a table;
        </li>
      </ul>
    </li>
    <li>
      Then there are the following modules and functions from Lua's standard
      library, all pre-imported and available:
      <ul>
        <li><code>pairs</code>;</li>
        <li><code>ipairs</code>;</li>
        <li><code>next</code>;</li>
        <li><code>error</code>;</li>
        <li><code>tonumber</code>;</li>
        <li><code>tostring</code>;</li>
        <li><code>type</code>;</li>
        <li><code>unpack</code>;</li>
        <li><code>string</code> with most its functions;</li>
        <li><code>table</code> with most its functions;</li>
        <li><code>math</code> with most its functions;</li>
        <li>
          <code>os</code> with <code>time</code>, <code>clock</code>,
          <code>difftime</code> and <code>date</code> functions;
        </li>
      </ul>
    </li>
  </ul>
  <h1 id="json-api">JSON API</h1>
  <p>
    Anything you can do on this website you can also do through Etleneum's
    public JSON API.
  </p>
  <h2>Types</h2>
  <ul>
    <li>
      <code>Contract</code>:
      <code
        >&#123;id: String, code: String, name: String, readme: String, funds:
        Int&#125;</code
      >
    </li>
    <li>
      <code>Call</code>:
      <code
        >&#123;id: String, time: String, method: String, payload: Any, matoshi:
        Int, cost: Int&#125;</code
      >
    </li>
  </ul>
  <h2>Endpoints</h2>
  <p>
    All paths start at <code>https://etleneum.com</code> and must be called with
    <code>Content-Type: application/json</code>. All methods are
    <a href="https://developer.mozilla.org/en-US/docs/Web/HTTP/CORS">CORS</a
    >-enabled and no authorization mechanism is required or supported.<br />
    All calls return an object of type
    <code>&#123;ok: Bool, error: String, value: Any&#125;</code>. The relevant
    data is always in the <code>value</code> key and <code>error</code> is only
    present when the call has failed. In the following endpoint descriptions we
    omit the <code>ok/value</code> envelope and show just what should be inside
    <code>value</code>.
  </p>
  <ul>
    <li>
      <code>GET</code> <code>/~/contracts</code> lists all the contracts, sorted
      by the most recent activity, returns <code>Contract</code>;
    </li>
    <li>
      <code>POST</code> <code>/~/contract</code> prepares a new contract, takes
      <code>&#123;name: String, code: String, readme: String&#125;</code>,
      returns <code>&#123;id: String, invoice: String&#125;</code>, when the
      invoice is paid the <code>__init__</code> call is executed and the contract
      is created;
    </li>
    <li>
      <code>GET</code> <code>/~/contract/&lt;id&gt;</code> returns the full
      contract info, <code>Contract</code>;
    </li>
    <li>
      <code>GET</code> <code>/~/contract/&lt;id&gt;/state</code> returns just
      the contract state, <code>Any</code>;
    </li>
    <li>
      <code>POST</code>
      <code>/~/contract/&lt;id&gt;/state</code> with &lt;jq_filter&gt; as the
      body, or
      <code>GET</code>
      <code>/~/contract/&lt;id&gt;/state/&lt;jq_filter&gt;</code>
      returns the contract state after a
      <a href="https://stedolan.github.io/jq/manual/">jq</a> filter has been
      applied to it, <code>Any</code>;
    </li>
    <li>
      <code>GET</code> <code>/~/contract/&lt;id&gt;/funds</code> returns just
      the contract funds, in msat, <code>Int</code>;
    </li>
    <li>
      <code>SSE</code> <code>/~~~/contract/&lt;id&gt;</code> returns a
      <code>text/event-stream</code> that emits the following events:
      <ul>
        <li><code>contract-created: &#123;id: String&#125;</code>;</li>
        <li>
          <code
            >contract-error: &#123;id: String, kind: "internal" | "runtime",
            message: String&#125;</code
          >;
        </li>
        <li>
          <code
            >call-run-event: &#123;id: String, contract_id: String, kind:
            "start" | "print" | "run" String, message: String, method: String,
            msatoshi?: Int&#125;</code
          >;
        </li>
        <li>
          <code
            >call-made: &#123;id: String, contract_id: String, method:
            String&#125;</code
          >;
        </li>
        <li>
          <code
            >call-error: &#123;id: String, contract_id: String, kind: "internal"
            | "runtime", message: String, method: String&#125;</code
          >;
        </li>
      </ul>
    </li>
    <li>
      <code>POST</code> <code>/~/contract/&lt;id&gt;/call</code> prepares a new
      call, takes
      <code>&#123;method: String, payload: Any, msatoshi: Int&#125;</code>,
      returns <code>&#123;id: String, invoice: String&#125;</code>, when the
      invoice is paid the call is executed;
    </li>
    <li>
      <code>GET</code>
      <code>/~/contract/&lt;id&gt;/call/&lt;id&gt;</code> returns the full call
      info, <code>Call</code>;
    </li>
    <li>
      <code>PATCH</code>
      <code>/~/contract/&lt;id&gt;/call/&lt;id&gt;</code> takes anything passed
      in the JSON body and patches it to the current prepared call
      <strong>payload</strong>, returns the full call info, <code>Call</code>;
    </li>
    <li>
      <code>GET</code> <code>/lnurl/auth</code> performs
      <a href="https://github.com/fiatjaf/lnurl-rfc/blob/luds/04.md"
        >lnurl-auth</a
      >
      and creates a session;
    </li>
    <li>
      <code>SSE</code> <code>/~~~/session[?session=...]</code> returns a
      <code>text/event-stream</code> that emits the following events:
      <ul>
        <li>
          <code>lnurls: &#123;auth: String, withdraw: String&#125;</code>;
        </li>
        <li>
          <code
            >auth: &#123;account: String, balance: Int, [secret: String]&#125;</code
          >;
        </li>
        <li>
          <code>withdraw: &#123;amount: Int, new_balance: Int&#125;</code>;
        </li>
      </ul>
    </li>
  </ul>
</main>

<style>
  code {
    padding: 1px 2px;
    background-color: var(--lightgrey);
  }
</style>
