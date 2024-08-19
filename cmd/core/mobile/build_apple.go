// Copyright 2015 The Go Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mobile

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"cogentcore.org/core/base/exec"
	"cogentcore.org/core/cmd/core/config"
	"cogentcore.org/core/cmd/core/rendericon"
	"github.com/jackmordaunt/icns/v2"
	"golang.org/x/tools/go/packages"
)

// GoAppleBuild builds the given package with the given bundle ID for the given iOS targets.
func GoAppleBuild(c *config.Config, pkg *packages.Package, targets []config.Platform) (map[string]bool, error) {
	src := pkg.PkgPath

	err := SetupMoltenFramework()
	if err != nil {
		return nil, err
	}

	infoplist := new(bytes.Buffer)
	if err := InfoplistTmpl.Execute(infoplist, InfoplistTmplData{
		BundleID:           c.ID,
		Name:               c.Name,
		Version:            c.Version,
		InfoString:         c.About,
		ShortVersionString: c.Version,
		IconFile:           "icon.icns",
	}); err != nil {
		return nil, err
	}

	// Detect the team ID
	teamID, err := DetectTeamID()
	if err != nil {
		return nil, err
	}

	projPbxproj := new(bytes.Buffer)
	if err := ProjPbxprojTmpl.Execute(projPbxproj, ProjPbxprojTmplData{
		TeamID: teamID,
	}); err != nil {
		return nil, err
	}

	files := []struct {
		name     string
		contents []byte
	}{
		{TmpDir + "/main.xcodeproj/project.pbxproj", projPbxproj.Bytes()},
		{TmpDir + "/main/Info.plist", infoplist.Bytes()},
		{TmpDir + "/main/Images.xcassets/AppIcon.appiconset/Contents.json", []byte(ContentsJSON)},
	}

	for _, file := range files {
		if err := exec.MkdirAll(filepath.Dir(file.name), 0755); err != nil {
			return nil, err
		}
		exec.PrintCmd(fmt.Sprintf("echo \"%s\" > %s", file.contents, file.name), nil)
		if err := os.WriteFile(file.name, file.contents, 0644); err != nil {
			return nil, err
		}
	}

	// We are using lipo tool to build multiarchitecture binaries.
	args := []string{"lipo", "-o", filepath.Join(TmpDir, "main/main"), "-create"}

	var nmpkgs map[string]bool
	builtArch := map[string]bool{}
	for _, t := range targets {
		// Only one binary per arch allowed
		// e.g. ios/arm64 + iossimulator/amd64
		if builtArch[t.Arch] {
			continue
		}
		builtArch[t.Arch] = true

		path := filepath.Join(TmpDir, t.OS, t.Arch)

		// Disable DWARF; see golang.org/issues/25148.
		if err := GoBuild(c, src, AppleEnv[t.String()], "-ldflags", "-w "+config.LinkerFlags(c), "-o="+path); err != nil {
			return nil, err
		}
		if nmpkgs == nil {
			var err error
			nmpkgs, err = ExtractPkgs(c, AppleNM, path)
			if err != nil {
				return nil, err
			}
		}
		args = append(args, path)

	}

	if err := exec.Run("xcrun", args...); err != nil {
		return nil, err
	}

	if err := AppleCopyAssets(c, pkg, TmpDir); err != nil {
		return nil, err
	}

	// Build and move the release build to the output directory.
	err = exec.Run("xcrun", "xcodebuild",
		"-configuration", "Release",
		"-project", TmpDir+"/main.xcodeproj",
		"-allowProvisioningUpdates",
		"DEVELOPMENT_TEAM="+teamID)
	if err != nil {
		return nil, err
	}

	inm := filepath.Join(TmpDir+"/build/Release-iphoneos/main.app", "icon.icns")
	fdsi, err := os.Create(inm)
	if err != nil {
		return nil, err
	}
	defer fdsi.Close()
	// 1024x1024 is the largest icon size on iOS
	// (for the App Store)
	sic, err := rendericon.Render(1024)
	if err != nil {
		return nil, err
	}
	err = icns.Encode(fdsi, sic)
	if err != nil {
		return nil, err
	}

	// TODO(jbd): Fallback to copying if renaming fails.
	err = os.MkdirAll(filepath.Join("bin", "ios"), 0777)
	if err != nil {
		return nil, err
	}
	output := filepath.Join("bin", "ios", c.Name+".app")
	exec.PrintCmd(fmt.Sprintf("mv %s %s", TmpDir+"/build/Release-iphoneos/main.app", output), nil)
	// if output already exists, remove.
	if err := exec.RemoveAll(output); err != nil {
		return nil, err
	}
	if err := os.Rename(TmpDir+"/build/Release-iphoneos/main.app", output); err != nil {
		return nil, err
	}

	// need to copy framework
	// TODO(kai): could do framework setup step here
	err = exec.Run("cp", "-r", "$HOME/Library/CogentCore/MoltenVK.framework", output)
	if err != nil {
		return nil, err
	}
	return nmpkgs, nil
}

// SetupMoltenFramework creates the MoltenVK.framework file in the
// user's library if it doesn't already exist.
func SetupMoltenFramework() error {
	hdir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting user home directory: %w", err)
	}
	gdir := filepath.Join(hdir, "Library", "Cogent Core")
	_, err = os.Stat(filepath.Join(gdir, "MoltenVK.framework"))
	if err == nil {
		// it already exists
		return nil
	}

	tmp, err := os.MkdirTemp("", "cogent-core-setup-ios-vulkan-")
	if err != nil {
		return err
	}
	err = exec.Major().SetDir(tmp).Run("git", "clone", "https://github.com/goki/vulkan_mac_deps")
	if err != nil {
		return err
	}

	err = exec.MkdirAll(gdir, 0750)
	if err != nil {
		return err
	}
	err = exec.Run("cp", "-r", filepath.Join(tmp, "vulkan_mac_deps", "sdk", "ios", "MoltenVK.framework"), gdir)
	if err != nil {
		return err
	}
	return exec.RemoveAll(tmp)
}

// DetectTeamID determines the Apple Development Team ID on the system.
func DetectTeamID() (string, error) {
	// Grabs the first certificate for "Apple Development"; will not work if there
	// are multiple certificates and the first is not desired.
	pemString, err := exec.Output(
		"security", "find-certificate",
		"-c", "Apple Development", "-p",
	)
	if err != nil {
		err = fmt.Errorf("failed to pull the signing certificate to determine your team ID: %v", err)
		return "", err
	}

	block, _ := pem.Decode([]byte(pemString))
	if block == nil {
		err = fmt.Errorf("failed to decode the PEM to determine your team ID: %s", pemString)
		return "", err
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		err = fmt.Errorf("failed to parse your signing certificate to determine your team ID: %v", err)
		return "", err
	}

	if len(cert.Subject.OrganizationalUnit) == 0 {
		err = fmt.Errorf("the signing certificate has no organizational unit (team ID)")
		return "", err
	}

	return cert.Subject.OrganizationalUnit[0], nil
}

func AppleCopyAssets(c *config.Config, pkg *packages.Package, xcodeProjDir string) error {
	dstAssets := xcodeProjDir + "/main/assets"
	return exec.MkdirAll(dstAssets, 0755)
}

type InfoplistTmplData struct {
	BundleID           string
	Name               string
	Version            string
	InfoString         string
	ShortVersionString string
	IconFile           string
}

var InfoplistTmpl = template.Must(template.New("infoplist").Parse(`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleDevelopmentRegion</key>
	<string>en</string>
	<key>CFBundleExecutable</key>
	<string>main</string>
	<key>CFBundleIdentifier</key>
	<string>{{.BundleID}}</string>
	<key>CFBundleInfoDictionaryVersion</key>
	<string>6.0</string>
	<key>CFBundleName</key>
	<string>{{.Name}}</string>
	<key>CFBundlePackageType</key>
	<string>APPL</string>
	<key>CFBundleSignature</key>
	<string>????</string>
	<key>CFBundleVersion</key>
	<string>{{ .Version }}</string>
	<key>CFBundleGetInfoString</key>
	<string>{{ .InfoString }}</string>
	<key>CFBundleShortVersionString</key>
	<string>{{ .ShortVersionString }}</string>
	<key>CFBundleIconFile</key>
	<string>{{ .IconFile }}</string>
	<key>LSRequiresIPhoneOS</key>
	<true/>
	<key>UILaunchStoryboardName</key>
	<string>LaunchScreen</string>
	<key>UIRequiredDeviceCapabilities</key>
	<array>
		<string>armv7</string>
	</array>
	<key>UISupportedInterfaceOrientations</key>
	<array>
		<string>UIInterfaceOrientationPortrait</string>
		<string>UIInterfaceOrientationLandscapeLeft</string>
		<string>UIInterfaceOrientationLandscapeRight</string>
	</array>
	<key>UISupportedInterfaceOrientations~ipad</key>
	<array>
		<string>UIInterfaceOrientationPortrait</string>
		<string>UIInterfaceOrientationPortraitUpsideDown</string>
		<string>UIInterfaceOrientationLandscapeLeft</string>
		<string>UIInterfaceOrientationLandscapeRight</string>
	</array>
</dict>
</plist>
`))

type ProjPbxprojTmplData struct {
	TeamID string
}

var ProjPbxprojTmpl = template.Must(template.New("projPbxproj").Parse(`// !$*UTF8*$!
{
	archiveVersion = 1;
	classes = {
	};
	objectVersion = 46;
	objects = {

/* Begin PBXBuildFile section */
		254BB84F1B1FD08900C56DE9 /* Images.xcassets in Resources */ = {isa = PBXBuildFile; fileRef = 254BB84E1B1FD08900C56DE9 /* Images.xcassets */; };
		254BB8681B1FD16500C56DE9 /* main in Resources */ = {isa = PBXBuildFile; fileRef = 254BB8671B1FD16500C56DE9 /* main */; };
		25FB30331B30FDEE0005924C /* assets in Resources */ = {isa = PBXBuildFile; fileRef = 25FB30321B30FDEE0005924C /* assets */; };
/* End PBXBuildFile section */

/* Begin PBXFileReference section */
		254BB83E1B1FD08900C56DE9 /* main.app */ = {isa = PBXFileReference; explicitFileType = wrapper.application; includeInIndex = 0; path = main.app; sourceTree = BUILT_PRODUCTS_DIR; };
		254BB8421B1FD08900C56DE9 /* Info.plist */ = {isa = PBXFileReference; lastKnownFileType = text.plist.xml; path = Info.plist; sourceTree = "<group>"; };
		254BB84E1B1FD08900C56DE9 /* Images.xcassets */ = {isa = PBXFileReference; lastKnownFileType = folder.assetcatalog; path = Images.xcassets; sourceTree = "<group>"; };
		254BB8671B1FD16500C56DE9 /* main */ = {isa = PBXFileReference; lastKnownFileType = "compiled.mach-o.executable"; path = main; sourceTree = "<group>"; };
		25FB30321B30FDEE0005924C /* assets */ = {isa = PBXFileReference; lastKnownFileType = folder; name = assets; path = main/assets; sourceTree = "<group>"; };
/* End PBXFileReference section */

/* Begin PBXGroup section */
		254BB8351B1FD08900C56DE9 = {
			isa = PBXGroup;
			children = (
				25FB30321B30FDEE0005924C /* assets */,
				254BB8401B1FD08900C56DE9 /* main */,
				254BB83F1B1FD08900C56DE9 /* Products */,
			);
			sourceTree = "<group>";
			usesTabs = 0;
		};
		254BB83F1B1FD08900C56DE9 /* Products */ = {
			isa = PBXGroup;
			children = (
				254BB83E1B1FD08900C56DE9 /* main.app */,
			);
			name = Products;
			sourceTree = "<group>";
		};
		254BB8401B1FD08900C56DE9 /* main */ = {
			isa = PBXGroup;
			children = (
				254BB8671B1FD16500C56DE9 /* main */,
				254BB84E1B1FD08900C56DE9 /* Images.xcassets */,
				254BB8411B1FD08900C56DE9 /* Supporting Files */,
			);
			path = main;
			sourceTree = "<group>";
		};
		254BB8411B1FD08900C56DE9 /* Supporting Files */ = {
			isa = PBXGroup;
			children = (
				254BB8421B1FD08900C56DE9 /* Info.plist */,
			);
			name = "Supporting Files";
			sourceTree = "<group>";
		};
/* End PBXGroup section */

/* Begin PBXNativeTarget section */
		254BB83D1B1FD08900C56DE9 /* main */ = {
			isa = PBXNativeTarget;
			buildConfigurationList = 254BB8611B1FD08900C56DE9 /* Build configuration list for PBXNativeTarget "main" */;
			buildPhases = (
				254BB83C1B1FD08900C56DE9 /* Resources */,
			);
			buildRules = (
			);
			dependencies = (
			);
			name = main;
			productName = main;
			productReference = 254BB83E1B1FD08900C56DE9 /* main.app */;
			productType = "com.apple.product-type.application";
		};
/* End PBXNativeTarget section */

/* Begin PBXProject section */
		254BB8361B1FD08900C56DE9 /* Project object */ = {
			isa = PBXProject;
			attributes = {
				LastUpgradeCheck = 0630;
				ORGANIZATIONNAME = Developer;
				TargetAttributes = {
					254BB83D1B1FD08900C56DE9 = {
						CreatedOnToolsVersion = 6.3.1;
						DevelopmentTeam = {{.TeamID}};
					};
				};
			};
			buildConfigurationList = 254BB8391B1FD08900C56DE9 /* Build configuration list for PBXProject "main" */;
			compatibilityVersion = "Xcode 3.2";
			developmentRegion = English;
			hasScannedForEncodings = 0;
			knownRegions = (
				en,
				Base,
			);
			mainGroup = 254BB8351B1FD08900C56DE9;
			productRefGroup = 254BB83F1B1FD08900C56DE9 /* Products */;
			projectDirPath = "";
			projectRoot = "";
			targets = (
				254BB83D1B1FD08900C56DE9 /* main */,
			);
		};
/* End PBXProject section */

/* Begin PBXResourcesBuildPhase section */
		254BB83C1B1FD08900C56DE9 /* Resources */ = {
			isa = PBXResourcesBuildPhase;
			buildActionMask = 2147483647;
			files = (
				25FB30331B30FDEE0005924C /* assets in Resources */,
				254BB8681B1FD16500C56DE9 /* main in Resources */,
				254BB84F1B1FD08900C56DE9 /* Images.xcassets in Resources */,
			);
			runOnlyForDeploymentPostprocessing = 0;
		};
/* End PBXResourcesBuildPhase section */

/* Begin XCBuildConfiguration section */
		254BB8601B1FD08900C56DE9 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ALWAYS_SEARCH_USER_PATHS = NO;
				CLANG_CXX_LANGUAGE_STANDARD = "gnu++0x";
				CLANG_CXX_LIBRARY = "libc++";
				CLANG_ENABLE_MODULES = YES;
				CLANG_ENABLE_OBJC_ARC = YES;
				CLANG_WARN_BOOL_CONVERSION = YES;
				CLANG_WARN_CONSTANT_CONVERSION = YES;
				CLANG_WARN_DIRECT_OBJC_ISA_USAGE = YES_ERROR;
				CLANG_WARN_EMPTY_BODY = YES;
				CLANG_WARN_ENUM_CONVERSION = YES;
				CLANG_WARN_INT_CONVERSION = YES;
				CLANG_WARN_OBJC_ROOT_CLASS = YES_ERROR;
				CLANG_WARN_UNREACHABLE_CODE = YES;
				CLANG_WARN__DUPLICATE_METHOD_MATCH = YES;
				"CODE_SIGN_IDENTITY[sdk=iphoneos*]" = "Apple Development";
				COPY_PHASE_STRIP = NO;
				DEBUG_INFORMATION_FORMAT = "dwarf-with-dsym";
				ENABLE_NS_ASSERTIONS = NO;
				ENABLE_STRICT_OBJC_MSGSEND = YES;
				GCC_C_LANGUAGE_STANDARD = gnu99;
				GCC_NO_COMMON_BLOCKS = YES;
				GCC_WARN_64_TO_32_BIT_CONVERSION = YES;
				GCC_WARN_ABOUT_RETURN_TYPE = YES_ERROR;
				GCC_WARN_UNDECLARED_SELECTOR = YES;
				GCC_WARN_UNINITIALIZED_AUTOS = YES_AGGRESSIVE;
				GCC_WARN_UNUSED_FUNCTION = YES;
				GCC_WARN_UNUSED_VARIABLE = YES;
				MTL_ENABLE_DEBUG_INFO = NO;
				SDKROOT = iphoneos;
				TARGETED_DEVICE_FAMILY = "1,2";
				VALIDATE_PRODUCT = YES;
				ENABLE_BITCODE = NO;
				IPHONEOS_DEPLOYMENT_TARGET = 15.0;
			};
			name = Release;
		};
		254BB8631B1FD08900C56DE9 /* Release */ = {
			isa = XCBuildConfiguration;
			buildSettings = {
				ASSETCATALOG_COMPILER_APPICON_NAME = AppIcon;
				INFOPLIST_FILE = main/Info.plist;
				LD_RUNPATH_SEARCH_PATHS = "$(inherited) @executable_path/Frameworks";
				PRODUCT_NAME = "$(TARGET_NAME)";
			};
			name = Release;
		};
/* End XCBuildConfiguration section */

/* Begin XCConfigurationList section */
		254BB8391B1FD08900C56DE9 /* Build configuration list for PBXProject "main" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				254BB8601B1FD08900C56DE9 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
		254BB8611B1FD08900C56DE9 /* Build configuration list for PBXNativeTarget "main" */ = {
			isa = XCConfigurationList;
			buildConfigurations = (
				254BB8631B1FD08900C56DE9 /* Release */,
			);
			defaultConfigurationIsVisible = 0;
			defaultConfigurationName = Release;
		};
/* End XCConfigurationList section */
	};
	rootObject = 254BB8361B1FD08900C56DE9 /* Project object */;
}
`))

const ContentsJSON = `{
	"images" : [
		{
			"idiom" : "iphone",
			"size" : "29x29",
			"scale" : "2x"
		},
		{
			"idiom" : "iphone",
			"size" : "29x29",
			"scale" : "3x"
		},
		{
			"idiom" : "iphone",
			"size" : "40x40",
			"scale" : "2x"
		},
		{
			"idiom" : "iphone",
			"size" : "40x40",
			"scale" : "3x"
		},
		{
			"idiom" : "iphone",
			"size" : "60x60",
			"scale" : "2x"
		},
		{
			"idiom" : "iphone",
			"size" : "60x60",
			"scale" : "3x"
		},
		{
			"idiom" : "ipad",
			"size" : "29x29",
			"scale" : "1x"
		},
		{
			"idiom" : "ipad",
			"size" : "29x29",
			"scale" : "2x"
		},
		{
			"idiom" : "ipad",
			"size" : "40x40",
			"scale" : "1x"
		},
		{
			"idiom" : "ipad",
			"size" : "40x40",
			"scale" : "2x"
		},
		{
			"idiom" : "ipad",
			"size" : "76x76",
			"scale" : "1x"
		},
		{
			"idiom" : "ipad",
			"size" : "76x76",
			"scale" : "2x"
		}
	],
	"info" : {
		"version" : 1,
		"author" : "xcode"
	}
}
`

// RFC1034Label sanitizes the name to be usable in a uniform type identifier.
// The sanitization is similar to xcode's rfc1034identifier macro that
// replaces illegal characters (not conforming the rfc1034 label rule) with '-'.
func RFC1034Label(name string) string {
	// * Uniform type identifier:
	//
	// According to
	// https://developer.apple.com/library/ios/documentation/FileManagement/Conceptual/understanding_utis/understand_utis_conc/understand_utis_conc.html
	//
	// A uniform type identifier is a Unicode string that usually contains characters
	// in the ASCII character set. However, only a subset of the ASCII characters are
	// permitted. You may use the Roman alphabet in upper and lower case (A–Z, a–z),
	// the digits 0 through 9, the dot (“.”), and the hyphen (“-”). This restriction
	// is based on DNS name restrictions, set forth in RFC 1035.
	//
	// Uniform type identifiers may also contain any of the Unicode characters greater
	// than U+007F.
	//
	// Note: the actual implementation of xcode does not allow some unicode characters
	// greater than U+007f. In this implementation, we just replace everything non
	// alphanumeric with "-" like the rfc1034identifier macro.
	//
	// * RFC1034 Label
	//
	// <label> ::= <letter> [ [ <ldh-str> ] <let-dig> ]
	// <ldh-str> ::= <let-dig-hyp> | <let-dig-hyp> <ldh-str>
	// <let-dig-hyp> ::= <let-dig> | "-"
	// <let-dig> ::= <letter> | <digit>
	const surrSelf = 0x10000
	begin := false

	var res []rune
	for i, r := range name {
		if r == '.' && !begin {
			continue
		}
		begin = true

		switch {
		case 'a' <= r && r <= 'z', 'A' <= r && r <= 'Z':
			res = append(res, r)
		case '0' <= r && r <= '9':
			if i == 0 {
				res = append(res, '-')
			} else {
				res = append(res, r)
			}
		default:
			if r < surrSelf {
				res = append(res, '-')
			} else {
				res = append(res, '-', '-')
			}
		}
	}
	return string(res)
}
