package main

func CmdSetup() {
	println("creating the required directories...")
	// create required directories
	for _, dir := range []string{
		LenvDir, SrcDir, LuaCfg.RootDir, LuaJitCfg.RootDir,
	} {
		printf("- %q: ", dir)
		if info, err := lstat(dir); err != nil {
			fatalf("cannot read metadata - %s", err)
		} else if info == nil {
			println("create")
			if err := mkdir(dir); err != nil {
				fatalf("failed to create - %s", err)
			}
		} else if info.IsDir() {
			println("found")
		} else {
			println("not a directory, please remove it yourself.")
		}
	}
}
