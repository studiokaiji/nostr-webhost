import { useEffect, useMemo } from "react";
import { H1 } from "./components/H1";
import { Link } from "./components/Link";

function App() {
  const htmlId = useMemo(() => window.location.pathname.split("/")[2], []);
  const cssId = useMemo(
    () => Array.from(document.styleSheets)[0].href?.split("/").slice(-1)[0],
    []
  );
  const jsId = useMemo(
    () => Array.from(document.scripts)[0].src?.split("/").slice(-1)[0],
    []
  );

  useEffect(() => {
    console.log(Array.from(document.styleSheets));
  }, []);

  return (
    <div className="font-body font-medium">
      <div className="h-screen relative w-full bg-primary text-white space-y-7 p-7">
        <H1 className="!leading-[0.8]">
          Nostr
          <br />
          WebHost
        </H1>
        <div>
          <p>
            Relay URL:{" "}
            <Link href="wss://nostrwebhost.studiokaiji.com">
              wss://nostrwebhost.studiokaiji.com
            </Link>
          </p>
          <p>HTML ID: {htmlId}</p>
          <p>CSS ID: {cssId}</p>
          <p>Javascript ID: {jsId}</p>
        </div>
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
        </div>
        <p className="absolute right-7 bottom-7">
          😩{"<"}Hmmm... I want to ride an ostrich...
        </p>
      </div>
    </div>
  );
}

export default App;
