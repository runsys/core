// Copyright (c) 2019, Cogent Core. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xyzcore

/*

// Embed2D embeds a 2D Viewport on a vertically oriented plane, using a texture.
// The embedded viewport contains its own 2D scenegraph and receives events, with
// mouse coordinates translated into the 3D plane space.  The full range of Cogent Core
// 2D elements can be embedded.
type Embed2D struct {
	Solid

	// the embedded viewport to display
	Viewport *EmbedViewport

	// overall scaling factor relative to an arbitrary but sensible default scale based on size of viewport -- increase to increase size of view
	Zoom float32

	// if true, will be resized to fit its contents during initialization (though it will never get smaller than original size specified at creation) -- this requires having a core.Layout element (or derivative, such as core.Frame) as the first and only child of the Viewport
	FitContent bool

	// original standardized 96 DPI size -- the original size specified on creation -- actual size is affected by device pixel ratio and resizing due to FitContent
	StdSize image.Point

	// original size scaled according to logical dpi
	DPISize image.Point
}

const (
	// FitContent is used as arg for NewEmbed2D to specify that plane should be resized
	// to fit content.
	FitContent = true

	// FixesSize is used as arg for NewEmbed2D to specify that plane should remain a
	// specified fixed size (using )
	FixedSize = false
)

// NewEmbed2D adds a new embedded 2D viewport of given name and nominal size
// according to the standard 96 dpi resolution (i.e., actual size is adjusted relative
// to that using window's current Logical DPI scaling).  If fitContent is true and
// first and only element in Viewport is a core.Layout, then it will be resized
// to fit content size (though no smaller than given size).
func NewEmbed2D(sc *Scene, parent tree.Node, name string, width, height int, fitContent bool) *Embed2D {
	em := parent.NewChild(TypeEmbed2D, name).(*Embed2D)
	em.Defaults(sc)
	em.StdSize = image.Point{width, height}
	em.Viewport = NewEmbedViewport(sc, em, name, width, height)
	em.Viewport.Fill = true
	em.FitContent = fitContent
	return em
}

func (em *Embed2D) Defaults(sc *Scene) {
	tm := sc.PlaneMesh2D()
	em.SetMesh(sc, tm)
	em.Solid.Defaults()
	em.Zoom = 1
	em.Mat.Bright = 2 // this is key for making e.g., a white background show up as white..
}

// ResizeToFit resizes viewport and texture to fit the content
func (em *Embed2D) ResizeToFit() error {
	initSz := em.Viewport.Scene.Viewport.LayState.Alloc.Size.ToPoint()
	vpsz := em.Viewport.PrefSize(initSz)
	vpsz.X = ints.MaxInt(em.DPISize.X, vpsz.X)
	vpsz.Y = ints.MaxInt(em.DPISize.Y, vpsz.Y)
	em.Viewport.Resize(vpsz)
	em.Viewport.FullRender2DTree()
	em.UploadViewTex(em.Viewport.Scene)
	return nil
}

// Resize resizes viewport and texture to given standardized 96 DPI size,
// which becomes the specified new size.
func (em *Embed2D) Resize(width, height int) {
	em.StdSize = image.Point{width, height}
	em.SetDPISize()
	em.Viewport.Resize(em.DPISize)
	em.Viewport.FullRender2DTree()
	em.UploadViewTex(em.Viewport.Scene)
}

// SetDPISize sets the DPI-adjusted size using LogicalDPI from window.
// Window must be non-nil.   Als
func (em *Embed2D) SetDPISize() {
	if em.Viewport.Win == nil {
		return
	}
	ldpi := em.Viewport.Win.LogicalDPI()
	scr := ldpi / 96.0
	em.Zoom = 1.0 / scr
	// fmt.Printf("init ldpi: %v  scr: %v\n", ldpi, scr)
	sz := em.StdSize
	sz.X = int(float32(sz.X) * scr)
	sz.Y = int(float32(sz.Y) * scr)
	em.DPISize = sz
}

func (em *Embed2D) Config(sc *Scene) {
	if sc.Win != nil && em.Viewport.Win == nil {
		em.Viewport.Win = sc.Win
		em.SetDPISize()
		if em.FitContent {
			em.ResizeToFit()
		} else {
			em.Viewport.Resize(em.DPISize)
			em.Viewport.FullRender2DTree()
			em.UploadViewTex(sc)
		}
	} else {
		em.Viewport.FullRender2DTree()
		em.UploadViewTex(sc)
	}
	err := em.Validate(sc)
	if err != nil {
		em.SetInvisible()
	}
	em.NodeBase.Config(sc)
}

// UploadViewTex uploads the viewport image to the texture
func (em *Embed2D) UploadViewTex(sc *Scene) {
	img := em.Viewport.Pixels
	var tx Texture
	var err error
	if em.Mat.TexPtr == nil {
		txname := "__Embed2D: " + em.Nm
		tx, err = sc.TextureByNameTry(txname)
		if err != nil {
			tx = &TextureBase{Nm: txname}
			sc.AddTexture(tx)
			tx.SetImage(img)
			em.Mat.SetTexture(sc, tx)
		} else {
			fmt.Printf("xyz.Embed2D: error: texture name conflict: %s\n", txname)
			em.Mat.SetTexture(sc, tx)
		}
	} else {
		tx = em.Mat.TexPtr
		tx.SetImage(img)
		sc.Phong.UpdateTextureName(tx.Name)
	}
}

// Validate checks that text has valid mesh and texture settings, etc
func (em *Embed2D) Validate() error {
	// todo: validate more stuff here
	return em.Solid.Validate(sc)
}

func (em *Embed2D) UpdateWorldMatrix(parWorld *math32.Matrix4) {
	em.PoseMu.Lock()
	defer em.PoseMu.Unlock()
	if em.Viewport != nil {
		sz := em.Viewport.Geom.Size
		sc := math32.Vec3(.006 * em.Zoom * float32(sz.X), .006 * em.Zoom * float32(sz.Y), em.Pose.Scale.Z)
		em.Pose.Matrix.SetTransform(em.Pose.Pos, em.Pose.Quat, sc)
	} else {
		em.Pose.UpdateMatrix()
	}
	em.Pose.UpdateWorldMatrix(parWorld)
	em.SetFlag(int(WorldMatrixUpdated))
}

func (em *Embed2D) UpdateBBox2D(size math32.Vector2, sc *Scene) {
	em.Solid.UpdateBBox2D(size, sc)
	em.Viewport.BBoxMu.Lock()
	em.BBoxMu.Lock()
	em.Viewport.Geom.Pos = em.WinBBox.Min
	em.Viewport.WinBBox.Min = em.WinBBox.Min
	em.Viewport.WinBBox.Max = em.WinBBox.Min.Add(em.Viewport.Geom.Size)
	em.BBoxMu.Unlock()
	em.Viewport.BBoxMu.Unlock()
	em.Viewport.FuncDownMeFirst(0, em.Viewport.This, func(k tree.Node, level int, d any) bool {
		if k == em.Viewport.This {
			return tree.Continue
		}
		_, ni := core.KiToNode2D(k)
		if ni == nil {
			return tree.Break // going into a different type of thing, bail
		}
		if ni.IsUpdating() {
			return tree.Break
		}
		ni.SetWinBBox()
		return tree.Continue
	})
}

func (em *Embed2D) RenderClass() RenderClasses {
	return RClassOpaqueTexture
}

var Embed2DProperties = tree.Properties{
	tree.EnumTypeFlag: core.TypeNodeFlags,
}

func (em *Embed2D) Project2D(sc *Scene, pt image.Point) (image.Point, bool) {
	var ppt image.Point
	if em.Viewport == nil || em.Viewport.Pixels == nil {
		return ppt, false
	}
	if em.IsUpdating() {
		return ppt, false
	}
	em.Viewport.BBoxMu.RLock()
	em.BBoxMu.RLock()
	defer func() {
		em.BBoxMu.RUnlock()
		em.Viewport.BBoxMu.RUnlock()
	}()
	sz := em.Viewport.Geom.Size
	relpos := pt.Sub(sc.ObjBBox.Min)
	ray := em.RayPick(relpos, sc)
	// is in XY plane with norm pointing up in Z axis
	plane := math32.Plane{Norm: math32.Vec3(0, 0, 1), Off: 0}
	ispt, ok := ray.IntersectPlane(plane)
	if !ok || ispt.Z > 1.0e-5 { // Z > 0 means clicked "in front" of plane -- with tolerance
		fmt.Printf("in front: ok: %v   ispt: %v\n", ok, ispt)
		return ppt, false
	}
	ppt.X = int((ispt.X + 0.5) * float32(sz.X))
	ppt.Y = int((ispt.Y + 0.5) * float32(sz.Y))
	if ppt.X < 0 || ppt.Y < 0 {
		return ppt, false
	}
	ppt.Y = sz.Y - ppt.Y // top-zero
	// fmt.Printf("ppt: %v\n", ppt)
	ppt = ppt.Add(em.Viewport.Geom.Pos)
	return ppt, true
}

func (em *Embed2D) ConnectEvents3D(sc *Scene) {
	em.ConnectEvent(sc.Win, system.MouseEvent, core.RegPri, func(recv, send tree.Node, sig int64, d any) {
		emm := recv.Embed(TypeEmbed2D).(*Embed2D)
		ssc := emm.Viewport.Scene
		if !ssc.IsVisible() {
			return
		}
		cpop := ssc.Win.CurPopup()
		if cpop != nil && !ssc.Win.CurPopupIsTooltip() {
			return // let window handle popups
		}
		me := d.(*mouse.Event)
		ppt, ok := emm.Project2D(ssc, me.Where)
		if !ok {
			return
		}
		if !ssc.HasFocus2D() {
			ssc.GrabFocus()
		}
		md := &mouse.Event{}
		*md = *me
		md.Where = ppt
		emm.Viewport.Events.MouseEvents(md)
		emm.Viewport.Events.SendEventSignal(md, false)
		emm.Viewport.Events.MouseEventReset(md)
		if !md.IsProcessed() {
			ni := em.This.(Node)
			if ssc.CurSel != ni {
				ssc.SetSel(ni)
			}
		}
		me.SetProcessed() // must always
	})
	em.ConnectEvent(sc.Win, system.MouseMoveEvent, core.RegPri, func(recv, send tree.Node, sig int64, d any) {
		emm := recv.Embed(TypeEmbed2D).(*Embed2D)
		ssc := emm.Viewport.Scene
		if !ssc.IsVisible() {
			return
		}
		cpop := ssc.Win.CurPopup()
		if cpop != nil && !ssc.Win.CurPopupIsTooltip() {
			return // let window handle popups
		}
		me := d.(*mouse.MoveEvent)
		ppt, ok := emm.Project2D(ssc, me.Where)
		if !ok {
			return
		}
		md := &mouse.MoveEvent{}
		*md = *me
		del := ppt.Sub(me.Where)
		md.Where = ppt
		md.From = md.From.Add(del)
		emm.Viewport.Events.MouseEvents(md)
		emm.Viewport.Events.SendEventSignal(md, false)
		emm.Viewport.Events.GenMouseFocusEvents(md, false)
		emm.Viewport.Events.MouseEventReset(md)
		me.SetProcessed() // must always
	})
	em.ConnectEvent(sc.Win, system.MouseDragEvent, core.RegPri, func(recv, send tree.Node, sig int64, d any) {
		emm := recv.Embed(TypeEmbed2D).(*Embed2D)
		ssc := emm.Viewport.Scene
		if !ssc.IsVisible() {
			return
		}
		cpop := ssc.Win.CurPopup()
		if cpop != nil && !ssc.Win.CurPopupIsTooltip() {
			return // let window handle popups
		}
		me := d.(*mouse.DragEvent)
		ppt, ok := emm.Project2D(ssc, me.Where)
		if !ok {
			return
		}
		md := &mouse.DragEvent{}
		*md = *me
		del := ppt.Sub(me.Where)
		md.Where = ppt
		md.From = md.From.Add(del)
		emm.Viewport.Events.MouseEvents(md)
		emm.Viewport.Events.SendEventSignal(md, false)
		emm.Viewport.Events.MouseEventReset(md)
		me.SetProcessed() // must always
	})
	em.ConnectEvent(sc.Win, system.KeyChordEvent, core.HiPri, func(recv, send tree.Node, sig int64, d any) {
		// note: restylesering HiPri -- we are outside 2D focus system, and get *all* keyboard events
		emm := recv.Embed(TypeEmbed2D).(*Embed2D)
		ssc := emm.Viewport.Scene
		if !ssc.IsVisible() || !ssc.HasFocus2D() {
			return
		}
		cpop := ssc.Win.CurPopup()
		if cpop != nil && !ssc.Win.CurPopupIsTooltip() {
			return // let window handle popups
		}
		kt := d.(*key.ChordEvent)
		// fmt.Printf("key event: %v\n", kt.String())
		emm.Viewport.Events.MouseEvents(kt) // also handles key..
		emm.Viewport.Events.SendEventSignal(kt, false)
		if !kt.IsProcessed() {
			emm.Viewport.Events.ManagerKeyChordEvents(kt)
		}
		emm.Viewport.Events.MouseEventReset(kt)
	})
}

///////////////////////////////////////////////////////////////////
//  EmbedViewport

// EmbedViewport is an embedded viewport with its own event manager to handle
// events instead of using the Window.
type EmbedViewport struct {
	core.Viewport2D

	// event manager that handles dispersing events to nodes
	Events core.Events `json:"-" xml:"-"`

	// parent scene -- trigger updates
	Scene *Scene `json:"-" xml:"-"`

	// parent Embed2D -- render updates
	EmbedPar *Embed2D `json:"-" xml:"-"`

	// update flag for top-level updates
	TopUpdated bool `json:"-" xml:"-"`
}

// NewEmbedViewport creates a new Pixels Image with the specified width and height,
// and initializes the renderer etc
func NewEmbedViewport(sc *Scene, em *Embed2D, name string, width, height int) *EmbedViewport {
	sz := image.Point{width, height}
	vp := &EmbedViewport{}
	vp.Geom = core.Geom2DInt{Size: sz}
	vp.Pixels = image.NewRGBA(image.Rectangle{Max: sz})
	vp.Render.Init(width, height, vp.Pixels)
	vp.InitName(vp, name)
	vp.Scene = sc
	vp.EmbedPar = em
	vp.Win = vp.Scene.Win
	vp.Events.Master = vp
	return vp
}

func (vp *EmbedViewport) VpTop() core.Viewport {
	return vp.This.(core.Viewport)
}

func (vp *EmbedViewport) VpTopNode() core.Node {
	return vp.This.(core.Node)
}

func (vp *EmbedViewport) VpTopUpdateStart() bool {
	if vp.TopUpdated {
		return false
	}
	vp.TopUpdated = true
	return true
}

func (vp *EmbedViewport) VpTopUpdateEnd(update bool) {
	if !update {
		return
	}
	vp.VpUploadAll()
	vp.TopUpdated = false
}

func (vp *EmbedViewport) VpEvents() *core.Events {
	return &vp.Events
}

func (vp *EmbedViewport) VpIsVisible() bool {
	if vp.Scene == nil || vp.EmbedPar == nil {
		return false
	}
	return vp.Scene.IsVisible()
}

func (vp *EmbedViewport) VpUploadAll() {
	if !vp.This.(core.Viewport).VpIsVisible() {
		return
	}
	// fmt.Printf("embed vp upload all\n")
	update := vp.Scene.UpdateStart()
	if update {
		ssc := vp.EmbedPar.Viewport.Scene
		if !ssc.IsRendering() {
			vp.EmbedPar.UploadViewTex(vp.Scene)
		}
	}
	vp.Scene.UpdateEnd(update)
}

// VpUploadVp uploads our viewport image into the parent window -- e.g., called
// by popups when updating separately
func (vp *EmbedViewport) VpUploadVp() {
	vp.VpUploadAll()
}

// VpUploadRegion uploads node region of our viewport image
func (vp *EmbedViewport) VpUploadRegion(vpBBox, winBBox image.Rectangle) {
	vp.VpUploadAll()
}

///////////////////////////////////////
//  EventMaster API

func (vp *EmbedViewport) EventTopNode() tree.Node {
	return vp
}

func (vp *EmbedViewport) FocusTopNode() tree.Node {
	return vp
}

func (vp *EmbedViewport) EventTopUpdateStart() bool {
	return vp.VpTopUpdateStart()
}

func (vp *EmbedViewport) EventTopUpdateEnd(update bool) {
	vp.VpTopUpdateEnd(update)
}

// IsInScope returns whether given node is in scope for receiving events
func (vp *EmbedViewport) IsInScope(node tree.Node, popup bool) bool {
	return true // no popups for embedded
}

func (vp *EmbedViewport) CurPopupIsTooltip() bool {
	return false
}

// DeleteTooltip deletes any tooltip popup (called when hover ends)
func (vp *EmbedViewport) DeleteTooltip() {

}

// IsFocusActive returns true if focus is active in this master
func (vp *EmbedViewport) IsFocusActive() bool {
	return true
}

// SetFocusActiveState sets focus active state
func (vp *EmbedViewport) SetFocusActiveState(active bool) {
}

*/
