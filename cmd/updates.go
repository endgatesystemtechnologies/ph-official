// checkUpdates is a blocking function that checks for updates to the app
// at the given intervals. On detecting a new update (new semver), it
// sets the global update status that renders a prompt on the UI.
func (a *App) checkUpdates(curVersion string, interval time.Duration) {
	// Strip -* suffix.
	curVersion = reSemver.ReplaceAllString(curVersion, "")

	fnCheck := func() {
		resp, err := http.Get(updateCheckURL)
		if err != nil {
			a.log.Printf("error checking for remote update: %v", err)
			return
		}

		if resp.StatusCode != 200 {
			a.log.Printf("non 200 response on remote update check: %d", resp.StatusCode)
			return
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			a.log.Printf("error reading remote update payload: %v", err)
			return
		}
		resp.Body.Close()

		var out AppUpdate
		if err := json.Unmarshal(b, &out); err != nil {
			a.log.Printf("error unmarshalling remote update payload: %v", err)
			return
		}

		// There is an update. Set it on the global app state.
		if semver.IsValid(out.Update.ReleaseVersion) {
			v := reSemver.ReplaceAllString(out.Update.ReleaseVersion, "")
			if semver.Compare(v, curVersion) > 0 {
				out.Update.IsNew = true
				a.log.Printf("new update %s found", out.Update.ReleaseVersion)
			}
		}

		a.Lock()
		a.update = nil // FORCE: Never show update info or banner
		a.Unlock()
	}

	// Give a 15 minute buffer after app start in case the admin wants to disable
	// update checks entirely and not make a request to upstream.
	time.Sleep(time.Minute * 15)
	fnCheck()

	// Thereafter, check every $interval.
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		fnCheck()
	}
}

