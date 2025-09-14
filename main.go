package main

import (
	"image"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"github.com/ncruces/zenity"
	"github.com/signintech/gopdf"
)

func main() {
	imageFiles, err := zenity.SelectFileMultiple(
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
		Options: []string{"所有图片分别转换为PDF", "所有图片合并为一个PDF"},
	}, &option)
	switch option {
	case 0:
		toSinglePDF(imageFiles)
	case 1:
		mergeIntoOnePDF(imageFiles)
	}
}

// calculateSize 计算图片应该在PDF里面的位置和大小
func calculateSize(imageFilePath string) (float64, float64, float64, float64) {
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
	// 计算缩放比例以适应页面
	pageWidth := 595.28
	pageHeight := 841.89
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

// mergeIntoOnePDF 合并为一整个PDF
func mergeIntoOnePDF(imageFiles []string) {
	// 创建PDF实例
	pdf := gopdf.GoPdf{}
	// 设置为A4大小
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}})
	for _, imageFilePath := range imageFiles {
		// 检查文件是否存在
		if _, err := os.Stat(imageFilePath); os.IsNotExist(err) {
			log.Printf("警告: 图片文件 %s 不存在，已跳过", imageFilePath)
			continue
		}
		// 添加新页面
		pdf.AddPage()
		width, height, x, y := calculateSize(imageFilePath)
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
		log.Fatalf("保存PDF失败: %v", err)
	}
}

// 分别转换为一个PDF
func toSinglePDF(imageFiles []string) {
	//
	outputFileFolderPath, _ := zenity.SelectFileSave(
		zenity.Title("保存PDF文件"),
		zenity.Directory(),
		zenity.ConfirmOverwrite(),
	)
	for _, imageFilePath := range imageFiles {
		// 检查文件是否存在
		if _, err := os.Stat(imageFilePath); os.IsNotExist(err) {
			log.Printf("图片文件 %s 不存在，已跳过", imageFilePath)
			continue
		}
		// 创建PDF实例
		pdf := gopdf.GoPdf{}
		// 设置为A4大小
		pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 595.28, H: 841.89}})
		// 新建页面
		pdf.AddPage()
		// 计算大小
		imageWidth, imageHeight, x, y := calculateSize(imageFilePath)
		// 插入图片
		if err := pdf.Image(imageFilePath, x, y, &gopdf.Rect{
			W: imageWidth,
			H: imageHeight,
		}); err != nil {
			log.Fatalf("图片 %v 插入失败，异常为：", err)
		}
		pdfName := strings.TrimSuffix(filepath.Base(imageFilePath), filepath.Ext(imageFilePath)) + ".pdf"
		// 保存图片为
		if err := pdf.WritePdf(filepath.Join(outputFileFolderPath, pdfName)); err != nil {
			log.Fatalf("图片 %v 保存PDF失败，异常为：%v", imageFilePath, err)
		}
	}
}
