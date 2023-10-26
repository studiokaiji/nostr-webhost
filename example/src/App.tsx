import { H1 } from "./components/H1";
import { Link } from "./components/Link";

function App() {
  return (
    <div className="font-body font-medium p-7">
      <div className="h-screen relative w-full text-white space-y-7">
        <H1 className="!leading-[0.8]">
          Nostr
          <br />
          WebHost
        </H1>
        <div>
          <p>
            GitHub:{" "}
            <Link href="https://github.com/studiokaiji/nostr-webhost">
              https://github.com/studiokaiji/nostr-webhost
            </Link>
          </p>
        </div>
        <div>
          <p>Created By @studiokaiji</p>
          <p>
            Nostr:
            npub194qhhn5vzzyrsqaugfms8c7ycqjyvhyguurra450nhlweatfzxkqy8tgkd
          </p>
          <p>
            GitHub:{" "}
            <Link href="https://github.com/studiokaiji">
              https://github.com/studiokaiji
            </Link>
          </p>
          <p>Lightning Address: floppystore07@walletofsatoshi.com</p>
        </div>

        <div className="text-6xl font-extrabold bg-rainbow-gradient bg-clip-text text-transparent">
          ↓↓ CUTE OSTRICHES GALLERY ↓↓
        </div>

        <div className="flex flex-wrap">
          <img
            src="/DALL·E Ostrich on Jamaican Beach.png"
            width={512}
            height={512}
          />
          <img src="/DALL·E Ostrich Hero.png" width={512} height={512} />
          <img
            src="/DALL·E Ostrich Illustration (1).png"
            width={512}
            height={512}
          />
          <img
            src="/DALL·E Ostrich Illustration (2).png"
            width={512}
            height={512}
          />
          <img
            src="/DALL·E Ostrich Illustration.png"
            width={512}
            height={512}
          />
        </div>
      </div>
    </div>
  );
}

export default App;
