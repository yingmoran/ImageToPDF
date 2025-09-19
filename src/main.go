package main

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/ncruces/zenity"
	"github.com/signintech/gopdf"
)

func main() {
	myApp := app.New()
	indexWindow := myApp.NewWindow("图片转PDF")
	// 标题标签
	titleLabel := widget.NewLabel("图片转PDF")
	// 创建一个占位布局，用来展示文件列表
	fileListContainer := container.NewVBox()
	// 选择图片按钮
	selectImageButton := widget.NewButton("选择图片", func() {
		imageFiles, _ := zenity.SelectFileMultiple(
			zenity.Title("选择你需要转换的图片"),
			zenity.Filename("."),
			zenity.FileFilters{
				{"图片文件", []string{"*.png", "*.gif", "*.ico", "*.jpg", "*.webp"}, true},
			},
		)
		for _, imageFilePath := range imageFiles {
			var imageInfoContainer *fyne.Container
			imageInfoContainer = container.NewHBox(
				widget.NewButton("↑", func() {
					i := getIndex(fileListContainer, imageInfoContainer)
					if i == 0 {
						return
					}
					var temp *fyne.Container
					temp = fileListContainer.Objects[i-1].(*fyne.Container)
					fileListContainer.Objects[i-1] = fileListContainer.Objects[i]
					fileListContainer.Objects[i] = temp
				}),
				widget.NewButton("↓", func() {
					i := getIndex(fileListContainer, imageInfoContainer)
					if i == len(fileListContainer.Objects)-1 {
						return
					}
					var temp *fyne.Container
					temp = fileListContainer.Objects[i+1].(*fyne.Container)
					fileListContainer.Objects[i+1] = fileListContainer.Objects[i]
					fileListContainer.Objects[i] = temp
				}),
				widget.NewLabel(imageFilePath),
				// 删除按钮，用来删除用不着的图片
				widget.NewButton("删除", func() {
					fileListContainer.Remove(imageInfoContainer)
				}),
			)
			fileListContainer.Add(imageInfoContainer)
		}
	})
	// 清空图片按钮
	deleteImageButton := widget.NewButton("清空图片", func() {
		fileListContainer.RemoveAll()
	})
	indexWindow.SetContent(container.NewVBox(
		// 标题
		container.New(layout.NewCenterLayout(), titleLabel),
		container.New(layout.NewCenterLayout(), container.NewHBox(
			// 选择图片按钮
			selectImageButton,
			// 清空图片按钮
			deleteImageButton,
		)),
		container.New(layout.NewCenterLayout(), fileListContainer),
		container.New(layout.NewCenterLayout(), container.NewHBox(
			widget.NewButton("批量转换为PDF", func() {
				imageFiles := getimageFilePath(fileListContainer)
				if len(imageFiles) == 0 {
					zenity.Warning("请选择至少一张图片", zenity.Title("警告"))
				} else {
					toSinglePDF(imageFiles)
				}
			}),
			widget.NewButton("合并为一个PDF", func() {
				imageFiles := getimageFilePath(fileListContainer)
				if len(imageFiles) == 0 {
					zenity.Warning("请选择至少一张图片", zenity.Title("警告"))
				} else {
					mergeIntoOnePDF(imageFiles)
				}
			}),
		)),
	))
	indexWindow.ShowAndRun()
	/*imageFiles, err := zenity.SelectFileMultiple(
		zenity.Title("选择你需要转换的图片"),
		zenity.Filename("."),
		zenity.FileFilters{
			{"图片文件", []string{"*.png", "*.gif", "*.ico", "*.jpg", "*.webp"}, true},
		},
	) // 初始目录
	if err != nil {
		log.Fatalf("文件选择对话框错误，异常为：%v", err)
	}
	option := 0
	survey.AskOne(&survey.Select{
		Message: "请选择你要执行的操作",
		Options: []string{"批量转换为PDF", "合并为一个PDF"},
	}, &option)
	switch option {
	case 0:
		// 分别转换为PDF
		toSinglePDF(imageFiles)
	case 1:
		// 合并为一个PDF
		mergeIntoOnePDF(imageFiles)
	}*/
}

func getIndex(fileListContainer *fyne.Container, imageInfoContainer *fyne.Container) int {
	for i, imageInfoContaineri := range fileListContainer.Objects {
		if imageInfoContainer == imageInfoContaineri {
			return i
		}
	}
	return 0
}

func getimageFilePath(fileListContainer *fyne.Container) []string {
	var imageFiles []string
	// 获取图片列表内的图片
	for _, imageOption := range fileListContainer.Objects {
		imageFilePath := imageOption.(*fyne.Container).Objects[2].(*widget.Label).Text
		imageFiles = append(imageFiles, imageFilePath)
	}
	return imageFiles
}

// calculateSize 计算图片应该在PDF里面的位置和大小
func calculateSize(imageFilePath string, size *gopdf.Rect) (float64, float64, float64, float64) {
	width, height := getImageInfo(imageFilePath)
	// 计算缩放比例以适应页面
	pageWidth := size.W
	pageHeight := size.H
	scale := 1.0
	if width > pageWidth {
		scale = pageWidth / width
	}
	if height*scale > pageHeight {
		scale = pageHeight / height
	}
	// 计算居中位置
	x := (pageWidth - width*scale) / 2
	y := (pageHeight - height*scale) / 2
	imageWidth := width * scale
	imageHeight := height * scale
	return imageWidth, imageHeight, x, y
}

// getImageInfo 获取图片信息
func getImageInfo(imageFilePath string) (float64, float64) {
	imageFile, err := os.Open(imageFilePath)
	defer imageFile.Close()
	// 解码图片以获取尺寸信息
	img, _, err := image.Decode(imageFile)
	if err != nil {
		log.Printf("无法解码图片: %v", err)
	}
	// 获取图片尺寸
	bounds := img.Bounds()
	// 获取图片的宽
	width := float64(bounds.Max.X)
	// 获取图片的高
	height := float64(bounds.Max.Y)
	return width, height
}

// mergeIntoOnePDF 合并为一整个PDF
func mergeIntoOnePDF(imageFiles []string) {
	// 创建PDF实例
	pdf := gopdf.GoPdf{}
	// size := sizeSelect(0)
	// 设置大小
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	for _, imageFilePath := range imageFiles {
		// 检查文件是否存在
		if _, err := os.Stat(imageFilePath); os.IsNotExist(err) {
			log.Printf("警告: 图片文件 %s 不存在，已跳过", imageFilePath)
			continue
		}
		// 添加新页面
		pdf.AddPage()
		width, height, x, y := calculateSize(imageFilePath, gopdf.PageSizeA4)
		// 插入图片
		if err := pdf.Image(imageFilePath, x, y, &gopdf.Rect{
			W: width,
			H: height,
		}); err != nil {
			log.Printf("插入图片失败 %s: %v，已跳过", imageFilePath, err)
		}
	}
	outputFilePath, err := zenity.SelectFileSave(
		zenity.Title("保存PDF文件"),
		zenity.Filename("未命名文件.pdf"),
		zenity.FileFilters{
			{"pdf文件", []string{"*.pdf"}, true},
		},
		zenity.ConfirmOverwrite(),
	)
	if err != nil {
		return
	}
	if err := pdf.WritePdf(outputFilePath); err != nil {
		zenity.Error("保存PDF失败")
		return
	}
	zenity.Info("保存PDF成功", zenity.Title("提示"))
}

// toSinglePDF 批量保存为PDF
func toSinglePDF(imageFiles []string) {
	// 获取图片尺寸
	// size := sizeSelect(1)
	// 设置PDF的保存路径
	outputFileFolderPath, err := zenity.SelectFileSave(
		zenity.Title("请设置PDF保存路径"),
		zenity.Directory(),
		zenity.ConfirmOverwrite(),
	)
	if err != nil {
		return
	}
	for _, imageFilePath := range imageFiles {
		// 检查文件是否存在
		if _, err := os.Stat(imageFilePath); os.IsNotExist(err) {
			// log.Printf("图片文件 %s 不存在，已跳过", imageFilePath)
			continue
		}
		// 创建PDF实例
		pdf := gopdf.GoPdf{}
		/*if size == nil {
			w, h := getImageInfo(imageFilePath)
			size = &gopdf.Rect{W: w, H: h}
		}*/
		// 设置大小
		pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
		// 新建页面
		pdf.AddPage()
		// 计算大小
		imageWidth, imageHeight, x, y := calculateSize(imageFilePath, gopdf.PageSizeA4)
		// 插入图片
		if err := pdf.Image(imageFilePath, x, y, &gopdf.Rect{
			W: imageWidth,
			H: imageHeight,
		}); err != nil {
			zenity.Error("图片 " + imageFilePath + "保存PDF失败")
			return
		}
		pdfName := strings.TrimSuffix(filepath.Base(imageFilePath), filepath.Ext(imageFilePath)) + ".pdf"
		// 保存图片为
		if err := pdf.WritePdf(filepath.Join(outputFileFolderPath, pdfName)); err != nil {
			zenity.Error("图片 " + imageFilePath + "保存PDF失败")
			return
		}
	}
	zenity.Info("批量保存PDF成功", zenity.Title("提示"))
}

// sizeSelect 选择PDF尺寸大小
/*func sizeSelect(state int) *gopdf.Rect {
	fmt.Println(state)
	options := []string{"A0(2384x3371)", "A1(1685x2384)", "A2(1190x1684)", "A3(842x1190)", "A4(595x842)", "A5(420x595)", "设置为图片尺寸", "自定义长宽"}
	if state != 1 {
		options = append(options[:len(options)-2], options[len(options)-2+1])
	}
	sizeIndex := ""
	survey.AskOne(&survey.Select{
		Message: "请选择页面尺寸(pt)",
		Options: options,
		Description: func(value string, index int) string {
			switch index {
			case 0:
				return "A0"
			case 1:
				return "A1"
			case 2:
				return "A2"
			case 3:
				return "A3"
			case 4:
				return "A4"
			case 5:
				return "A5"
			default:
				return value
			}
		},
	}, &sizeIndex)
	var size *gopdf.Rect
	switch sizeIndex {
	case "A0":
		size = gopdf.PageSizeA0
	case "A1":
		size = gopdf.PageSizeA1
	case "A2":
		size = gopdf.PageSizeA2
	case "A3":
		size = gopdf.PageSizeA3
	case "A4":
		size = gopdf.PageSizeA4
	case "A5":
		size = gopdf.PageSizeA5
	case "设置为图片尺寸":
		size = nil
	default: // 如果不是上面列出来的编号，就是自定义长款
		var w float64
		var h float64
		survey.AskOne(&survey.Input{
			Message: "请输入宽(pt)",
		}, &w)
		survey.AskOne(&survey.Input{
			Message: "请输入高(pt)",
		}, &h)
		size = &gopdf.Rect{W: w, H: h}
	}
	return size
}
*/
