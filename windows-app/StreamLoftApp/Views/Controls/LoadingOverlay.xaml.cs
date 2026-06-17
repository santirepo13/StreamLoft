using System;
using System.IO;
using System.Windows;
using System.Windows.Controls;

namespace StreamLoftApp.Views.Controls;

public partial class LoadingOverlay : UserControl
{
    public new static readonly DependencyProperty IsVisibleProperty =
        DependencyProperty.Register(nameof(IsVisible), typeof(bool), typeof(LoadingOverlay),
            new PropertyMetadata(false, OnIsVisibleChanged));

    public static readonly DependencyProperty MessageProperty =
        DependencyProperty.Register(nameof(Message), typeof(string), typeof(LoadingOverlay),
            new PropertyMetadata("Loading..."));

    private bool _initialized;

    public new bool IsVisible
    {
        get => (bool)GetValue(IsVisibleProperty);
        set => SetValue(IsVisibleProperty, value);
    }

    public string Message
    {
        get => (string)GetValue(MessageProperty);
        set => SetValue(MessageProperty, value);
    }

    public LoadingOverlay()
    {
        InitializeComponent();
        Browser.DefaultBackgroundColor = System.Drawing.Color.Transparent;
        Loaded += OnLoaded;
    }

    private async void OnLoaded(object sender, RoutedEventArgs e)
    {
        if (_initialized) return;
        _initialized = true;

        try
        {
            var jsonPath = System.IO.Path.Combine(
                AppDomain.CurrentDomain.BaseDirectory,
                "Views", "assets", "loading.json");

            if (!File.Exists(jsonPath))
                return;

            var json = File.ReadAllText(jsonPath);

            await Browser.EnsureCoreWebView2Async();

            var html = BuildHtml(json);
            Browser.NavigateToString(html);
        }
        catch
        {
        }
    }

    private static string BuildHtml(string animationJson)
    {
        return $@"<!DOCTYPE html>
<html>
<head>
<style>
*{{margin:0;padding:0;overflow:hidden}}
body{{background:transparent}}
#l{{width:100px;height:100px}}
</style>
</head>
<body>
<div id='l'></div>
<script src='https://unpkg.com/lottie-web@5.12.2/build/player/lottie.min.js'></script>
<script>
lottie.loadAnimation({{
    container:document.getElementById('l'),
    renderer:'svg',
    loop:true,
    autoplay:true,
    animationData:{animationJson}
}});
</script>
</body>
</html>";
    }

    private static void OnIsVisibleChanged(DependencyObject d, DependencyPropertyChangedEventArgs e)
    {
    }
}
